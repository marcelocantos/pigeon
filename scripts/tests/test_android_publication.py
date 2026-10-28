# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
import importlib.util
import io
from pathlib import Path
import unittest
from unittest.mock import patch
import zipfile

spec = importlib.util.spec_from_file_location(
    "publication", Path(__file__).resolve().parents[1] / "check-android-publication.py"
)
publication = importlib.util.module_from_spec(spec)
spec.loader.exec_module(publication)
SHA = "a" * 40


def zipped(entries):
    buffer = io.BytesIO()
    with zipfile.ZipFile(buffer, "w") as archive:
        for name, body in entries.items():
            archive.writestr(name, body)
    return buffer.getvalue()


class PublicationTest(unittest.TestCase):
    def setUp(self):
        self.pom = (f"<project><groupId>com.github.marcelocantos.pigeon</groupId>"
                    f"<artifactId>pigeon</artifactId><version>{SHA}</version>"
                    "<packaging>aar</packaging></project>").encode()
        self.jar = zipped({"com/marcelocantos/pigeon/crypto/E2EChannel.class": b"class"})
        self.aar = zipped({"jni/arm64-v8a/libpigeon-jni.so": b"\x7fELF" * 25001,
                           "classes.jar": self.jar})

    def run_check(self, pom=None, aar=None):
        def fetch(url, limit):
            return (self.pom if pom is None else pom) if url.endswith(".pom") else (self.aar if aar is None else aar)
        with patch.object(publication, "fetch", side_effect=fetch) as request:
            publication.check(SHA)
            self.assertEqual(request.call_count, 2)

    def test_complete_publication(self):
        self.run_check()

    def test_missing_native_library_fails(self):
        with self.assertRaisesRegex(RuntimeError, "jni/arm64-v8a"):
            self.run_check(aar=zipped({"classes.jar": self.jar}))

    def test_missing_kotlin_class_fails(self):
        with self.assertRaisesRegex(RuntimeError, "E2EChannel"):
            self.run_check(aar=zipped({"jni/arm64-v8a/libpigeon-jni.so": b"x" * 100001,
                                       "classes.jar": zipped({"unrelated.class": b"x"})}))

    def test_wrong_pom_coordinates_fail(self):
        with self.assertRaisesRegex(RuntimeError, "groupId"):
            self.run_check(pom=self.pom.replace(b"com.github.marcelocantos.pigeon", b"wrong.group"))

    def test_old_tag_rejected(self):
        with self.assertRaisesRegex(ValueError, "full 40-character"):
            publication.check("v0.34.0")


if __name__ == "__main__":
    unittest.main()
