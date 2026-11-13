package e2e

import (
	"fmt"
	"log"

	"github.com/konveyor/crane/e2e-tests/config"
	. "github.com/konveyor/crane/e2e-tests/framework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	// The two same-size payloads are deliberately the same length so a
	// tampered destination file differs only in content. The test asserts
	// the lengths match before relying on that property.
	checksumVerifySourceContent   = "checksum-source-payload-0123456789"
	checksumVerifyTamperedContent = "checksum-tamper-payload-0123456789"
	checksumVerifyOtherContent    = "checksum-other-source-payload"
	checksumVerifyOtherTampered   = "short"

	// Fixed mtimes let the test restore the destination mtime exactly while
	// corrupting only the file content, exercising rsync --checksum instead
	// of its default size-and-mtime quick check.
	checksumVerifyMtime      = "202001020304.05"
	checksumVerifyOtherMtime = "202001020304.06"
)

var _ = Describe("transfer-pvc checksum verification", func() {
	It("[MTA-8073] corrects destination content that matches source size and mtime when --verify is set",
		Label("tier0", "pvc-transfer", "direct"), func() {
			const (
				appName       = "pvc"
				namespace     = "mta-8073-verify"
				pvcName       = "data-pvc"
				seedPodName   = "mta-8073-seed"
				verifyPodName = "mta-8073-verify"
				reverifyPod   = "mta-8073-reverify"
				sameSizeFile  = "same-size.txt"
				otherFile     = "other.txt"
			)

			Expect(checksumVerifyTamperedContent).To(HaveLen(len(checksumVerifySourceContent)),
				"tampered payload must keep the source file size")
			Expect(checksumVerifyOtherTampered).NotTo(HaveLen(len(checksumVerifyOtherContent)),
				"the control payload must use a different size")

			scenario := NewMigrationScenario(
				appName,
				namespace,
				config.K8sDeployBin,
				config.CraneBin,
				config.SourceContext,
				config.TargetContext,
			)
			srcApp := scenario.SrcAppNonAdmin
			tgtApp := scenario.TgtAppNonAdmin
			srcApp.ExtraVars = map[string]any{
				"non_admin_user": "true",
				"pvc_name":       pvcName,
			}
			tgtApp.ExtraVars = map[string]any{
				"non_admin_user": "true",
				"pvc_name":       pvcName,
			}

			By("Grant namespace-admin permissions to non-admin users on source and target")
			kubectlSrc, kubectlTgt, cleanup, err := SetupActiveKubectlRunners(scenario, namespace)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				for _, k := range []KubectlRunner{scenario.KubectlSrc, scenario.KubectlTgt} {
					if _, err := k.Run("delete", "namespace", namespace, "--ignore-not-found=true", "--wait=true"); err != nil {
						log.Printf("cleanup: failed to delete namespace %q on context %q: %v", namespace, k.Context, err)
					}
				}
			})
			DeferCleanup(cleanup)

			runner := scenario.CraneNonAdmin
			paths, err := NewScenarioPaths("crane-mta-8073-*")
			Expect(err).NotTo(HaveOccurred())
			runner.WorkDir = paths.TempDir
			DeferCleanup(func() {
				if err := CleanupScenario(paths.TempDir, srcApp, tgtApp); err != nil {
					log.Printf("cleanup apps/tempdir: %v", err)
				}
			})

			By("Deploy an unattached source PVC")
			Expect(srcApp.Deploy()).NotTo(HaveOccurred())
			_, err = kubectlSrc.Run("get", "pvc", pvcName, "-n", namespace)
			Expect(err).NotTo(HaveOccurred())

			By("Seed source files and pin their modification times")
			Expect(kubectlSrc.ApplyYAMLSpec(SeedPodManifest(namespace, seedPodName, pvcName), namespace)).NotTo(HaveOccurred())
			DeferCleanup(func() {
				if _, err := kubectlSrc.Run("delete", "pod", seedPodName, "-n", namespace, "--ignore-not-found=true", "--wait=true"); err != nil {
					log.Printf("cleanup seed pod %q: %v", seedPodName, err)
				}
			})
			_, err = kubectlSrc.Run("wait", "--for=condition=Ready", "pod/"+seedPodName, "-n", namespace, "--timeout=120s")
			Expect(err).NotTo(HaveOccurred())

			seedCmd := fmt.Sprintf(
				"set -eu; printf '%%s' '%s' > /data/%s; touch -t %s /data/%s; "+
					"printf '%%s' '%s' > /data/%s; touch -t %s /data/%s",
				checksumVerifySourceContent, sameSizeFile, checksumVerifyMtime, sameSizeFile,
				checksumVerifyOtherContent, otherFile, checksumVerifyOtherMtime, otherFile,
			)
			_, err = kubectlSrc.Run("exec", seedPodName, "-n", namespace, "--", "/bin/sh", "-c", seedCmd)
			Expect(err).NotTo(HaveOccurred())

			srcSameSizeMD5, err := MD5SumFile(kubectlSrc, namespace, seedPodName, "/data/"+sameSizeFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(srcSameSizeMD5).NotTo(BeEmpty())
			srcOtherMD5, err := MD5SumFile(kubectlSrc, namespace, seedPodName, "/data/"+otherFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(srcOtherMD5).NotTo(BeEmpty())

			srcSize, srcMtime, err := FileSizeAndMtime(kubectlSrc, namespace, seedPodName, "/data/"+sameSizeFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(srcSize).To(Equal(int64(len(checksumVerifySourceContent))),
				"source file should have the seeded size")
			log.Printf("source %s stat (size %d mtime %d), md5: %s", sameSizeFile, srcSize, srcMtime, srcSameSizeMD5)

			By("Release the source PVC before transferring")
			_, err = kubectlSrc.Run("delete", "pod", seedPodName, "-n", namespace, "--wait=true", "--timeout=120s")
			Expect(err).NotTo(HaveOccurred())

			targetIP, err := GetClusterNodeIP(scenario.TgtApp.Context)
			Expect(err).NotTo(HaveOccurred())
			transferOpts := TransferPVCOptions{
				SourceContext:       srcApp.Context,
				TargetContext:       tgtApp.Context,
				PVCName:             pvcName,
				PVCNamespaceMap:     fmt.Sprintf("%s:%s", namespace, namespace),
				Subdomain:           fmt.Sprintf("%s.%s.%s.nip.io", pvcName, namespace, targetIP),
				DisableCloudStorage: true,
				Verify:              true,
			}

			By("Transfer the PVC with checksum verification enabled")
			output, err := runner.TransferPVCWithOutput(transferOpts)
			Expect(err).NotTo(HaveOccurred())
			log.Printf("first transfer-pvc output:\n%s", output)

			By("Verify the first transfer produced destination data matching the source")
			Expect(kubectlTgt.ApplyYAMLSpec(VerifyPodManifest(namespace, verifyPodName, pvcName), namespace)).NotTo(HaveOccurred())
			DeferCleanup(func() {
				if _, err := kubectlTgt.Run("delete", "pod", verifyPodName, "-n", namespace, "--ignore-not-found=true", "--wait=true"); err != nil {
					log.Printf("cleanup verify pod %q: %v", verifyPodName, err)
				}
			})
			_, err = kubectlTgt.Run("wait", "--for=condition=Ready", "pod/"+verifyPodName, "-n", namespace, "--timeout=120s")
			Expect(err).NotTo(HaveOccurred())

			tgtSameSizeMD5, err := MD5SumFile(kubectlTgt, namespace, verifyPodName, "/data/"+sameSizeFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(tgtSameSizeMD5).To(Equal(srcSameSizeMD5), "first transfer should copy identical content")
			tgtOtherMD5, err := MD5SumFile(kubectlTgt, namespace, verifyPodName, "/data/"+otherFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(tgtOtherMD5).To(Equal(srcOtherMD5), "first transfer should copy identical content")

			tgtSize, tgtMtime, err := FileSizeAndMtime(kubectlTgt, namespace, verifyPodName, "/data/"+sameSizeFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(tgtSize).To(Equal(srcSize), "first transfer should preserve the source size")
			Expect(tgtMtime).To(Equal(srcMtime), "first transfer should preserve the source mtime")

			By("Corrupt a destination file while preserving its size and modification time")
			tamperCmd := fmt.Sprintf(
				"set -eu; printf '%%s' '%s' > /data/%s; touch -t %s /data/%s; "+
					"printf '%%s' '%s' > /data/%s; touch -t %s /data/%s",
				checksumVerifyTamperedContent, sameSizeFile, checksumVerifyMtime, sameSizeFile,
				checksumVerifyOtherTampered, otherFile, checksumVerifyOtherMtime, otherFile,
			)
			_, err = kubectlTgt.Run("exec", verifyPodName, "-n", namespace, "--", "/bin/sh", "-c", tamperCmd)
			Expect(err).NotTo(HaveOccurred())

			tamperedSize, tamperedMtime, err := FileSizeAndMtime(kubectlTgt, namespace, verifyPodName, "/data/"+sameSizeFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(tamperedSize).To(Equal(srcSize),
				"tampered destination file must keep the source size")
			Expect(tamperedMtime).To(Equal(srcMtime),
				"tampered destination file must keep the source mtime")
			tamperedMD5, err := MD5SumFile(kubectlTgt, namespace, verifyPodName, "/data/"+sameSizeFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(tamperedMD5).NotTo(Equal(srcSameSizeMD5),
				"tampered destination file must differ from the source content")

			tamperedOtherSize, _, err := FileSizeAndMtime(kubectlTgt, namespace, verifyPodName, "/data/"+otherFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(tamperedOtherSize).To(Equal(int64(len(checksumVerifyOtherTampered))),
				"control file should be corrupted to a different size")

			By("Release the destination PVC before re-running the verified transfer")
			_, err = kubectlTgt.Run("delete", "pod", verifyPodName, "-n", namespace, "--wait=true", "--timeout=120s")
			Expect(err).NotTo(HaveOccurred())

			By("Wait for the first transfer's helpers to release the destination PVC")
			AssertNoTransferPVCLeftovers(kubectlSrc, []string{namespace}, pvcName)
			AssertNoTransferPVCLeftovers(kubectlTgt, []string{namespace}, pvcName)

			By("Re-run transfer-pvc with --verify and confirm it repairs the tampered data")
			output, err = runner.TransferPVCWithOutput(transferOpts)
			Expect(err).NotTo(HaveOccurred())
			log.Printf("second transfer-pvc output:\n%s", output)

			Expect(kubectlTgt.ApplyYAMLSpec(VerifyPodManifest(namespace, reverifyPod, pvcName), namespace)).NotTo(HaveOccurred())
			DeferCleanup(func() {
				if _, err := kubectlTgt.Run("delete", "pod", reverifyPod, "-n", namespace, "--ignore-not-found=true", "--wait=true"); err != nil {
					log.Printf("cleanup verify pod %q: %v", reverifyPod, err)
				}
			})
			_, err = kubectlTgt.Run("wait", "--for=condition=Ready", "pod/"+reverifyPod, "-n", namespace, "--timeout=120s")
			Expect(err).NotTo(HaveOccurred())

			for _, file := range []struct {
				name    string
				wantMD5 string
			}{
				{name: sameSizeFile, wantMD5: srcSameSizeMD5},
				{name: otherFile, wantMD5: srcOtherMD5},
			} {
				md5, err := MD5SumFile(kubectlTgt, namespace, reverifyPod, "/data/"+file.name)
				Expect(err).NotTo(HaveOccurred())
				Expect(md5).To(Equal(file.wantMD5),
					"destination file %s should match the source after --verify transfer", file.name)
			}

			finalSize, finalMtime, err := FileSizeAndMtime(kubectlTgt, namespace, reverifyPod, "/data/"+sameSizeFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(finalSize).To(Equal(srcSize), "verified transfer should preserve the source size")
			Expect(finalMtime).To(Equal(srcMtime), "verified transfer should preserve the source mtime")
		})
})
