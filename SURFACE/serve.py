"""Loopback, read-only access to one integrity-verified portable edition."""

import argparse
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from urllib.parse import parse_qs, unquote, urlsplit

from jsonschema import Draft202012Validator, ValidationError

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "SYSTEM"))
from verify import CEILING, binding, check_export, encode, load_json


def selected(index, ref):
    for item in index["resources"]:
        if item["ref"] == ref:
            return item
    raise KeyError(ref)


def read_resource(index, files, request):
    item = selected(index, request["resource"])
    return item["media_type"], files[item["path"]]


def resolve_reference(index, files, request):
    ref = request["ref"]
    target = index["references"][ref]
    return "application/json", encode({"revision": index["revision"], "ref": ref, **target})


def check_resource(index, files, request):
    item = selected(index, request["resource"])
    actual = binding(files[item["path"]])
    expected = {key: item[key] for key in ("sha256", "byte_length")}
    return "application/json", encode({
        "revision": index["revision"], "resource": item["ref"], **actual,
        "result": "MATCH" if actual == expected else "MISMATCH",
        "conformance": "NOT_EVALUATED", "effect_ceiling": CEILING,
    })


HANDLERS = {"read": read_resource, "resolve": resolve_reference, "check": check_resource}


def handler_for(files):
    index = load_json(files["index.json"])
    operations = {op["route"]: op for op in index["operations"]}
    media = {item["path"]: item["media_type"] for item in index["resources"]}
    media.update({"index.html": "text/html; charset=utf-8",
                  "index.json": "application/json", "manifest.json": "application/json"})

    class Handler(BaseHTTPRequestHandler):
        def setup(self):
            super().setup()
            self.connection.settimeout(10)

        def reply(self, status, media_type, data):
            self.send_response(status)
            self.send_header("Content-Type", media_type)
            self.send_header("Content-Length", str(len(data)))
            self.send_header("X-Content-Type-Options", "nosniff")
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Security-Policy",
                             "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; "
                             "connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'")
            self.end_headers()
            self.wfile.write(data)

        def do_GET(self):
            try:
                if len(self.path) > 2048:
                    raise ValueError("Request target too long")
                url = urlsplit(self.path)
                if url.scheme or url.netloc or url.fragment:
                    raise ValueError("Only local request targets are accepted")
                path = unquote(url.path, errors="strict")
                if path in operations:
                    op = operations[path]
                    query = parse_qs(url.query, keep_blank_values=True, strict_parsing=True,
                                     max_num_fields=1, errors="strict")
                    request = {key: values[0] for key, values in query.items()}
                    Draft202012Validator(op["input"]).validate(request)
                    media_type, data = HANDLERS[op["id"]](index, files, request)
                    if op["id"] != "read":
                        Draft202012Validator(op["output"]).validate(load_json(data))
                    self.reply(200, media_type, data)
                else:
                    if url.query:
                        raise ValueError("Static resources do not accept parameters")
                    name = "index.html" if path == "/" else path.removeprefix("/")
                    if not path.startswith("/") or name not in media:
                        raise KeyError(path)
                    self.reply(200, media[name], files[name])
            except (ValueError, ValidationError, UnicodeError):
                self.reply(400, "application/json", encode({"error": "Invalid bounded request"}))
            except KeyError:
                self.reply(404, "application/json", encode({"error": "Not exposed in this edition"}))

    return Handler


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--edition", type=Path, required=True)
    parser.add_argument("--manifest-sha256", required=True)
    parser.add_argument("--port", type=int, default=8765)
    args = parser.parse_args()
    try:
        if not 1 <= args.port <= 65535:
            raise ValueError("Port must be between 1 and 65535")
        files, report = check_export(args.edition, args.manifest_sha256)
        # Requests use this checked in-memory snapshot, not changing disk contents.
        with HTTPServer(("127.0.0.1", args.port), handler_for(files)) as server:
            print(f"Edition {report['revision']} at http://127.0.0.1:{args.port}/", flush=True)
            server.serve_forever()
    except (ValueError, KeyError, TypeError, OSError) as error:
        print(f"Cannot serve edition: {error}", file=sys.stderr)
        return 1
    except KeyboardInterrupt:
        return 0


if __name__ == "__main__":
    sys.exit(main())
