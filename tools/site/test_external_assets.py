import base64
import unittest
from html.parser import HTMLParser
from pathlib import Path
from urllib.parse import urlparse


SITE = Path(__file__).resolve().parents[2] / "site"


class ExternalAssetParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.assets = []

    def handle_starttag(self, tag, attrs):
        attributes = dict(attrs)
        url = attributes.get("src" if tag == "script" else "href", "")
        if not url.startswith("https://"):
            return
        if tag == "script" or (tag == "link" and urlparse(url).hostname == "cdnjs.cloudflare.com"):
            self.assets.append((url, attributes))


class ExternalAssetTests(unittest.TestCase):
    def test_executable_remote_assets_have_integrity(self):
        for page in (SITE / "index.html", SITE / "examples" / "index.html"):
            with self.subTest(page=page):
                parser = ExternalAssetParser()
                parser.feed(page.read_text(encoding="utf-8"))
                for url, attributes in parser.assets:
                    with self.subTest(url=url):
                        self.assertEqual(attributes.get("crossorigin"), "anonymous")
                        integrity = attributes.get("integrity", "")
                        self.assertTrue(integrity.startswith("sha384-"))
                        digest = base64.b64decode(integrity.removeprefix("sha384-"), validate=True)
                        self.assertEqual(len(digest), 48)


if __name__ == "__main__":
    unittest.main()
