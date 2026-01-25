import json
import sys
import os
import re
import time


def load_data_js(filepath):
    if not os.path.exists(filepath):
        return {"lastUpdate": 0, "repoUrl": "https://github.com/casbin/casbin", "entries": {}}

    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()

    # Strip window.BENCHMARK_DATA =
    match = re.search(r"window\.BENCHMARK_DATA\s*=\s*({.*});?", content, re.DOTALL)
    if match:
        try:
            return json.loads(match.group(1))
        except json.JSONDecodeError:
            print("Error decoding JSON from data.js", file=sys.stderr)
            sys.exit(1)
    return {"lastUpdate": 0, "repoUrl": "https://github.com/casbin/casbin", "entries": {}}


def save_data_js(filepath, data):
    content = f"window.BENCHMARK_DATA = {json.dumps(data, indent=4)};"
    with open(filepath, "w", encoding="utf-8") as f:
        f.write(content)


def main():
    if len(sys.argv) < 3:
        print("Usage: python update_data_js.py benchmark_result.json data.js")
        sys.exit(1)

    bench_file = sys.argv[1]
    data_js_file = sys.argv[2]

    with open(bench_file, "r", encoding="utf-8") as f:
        bench_data = json.load(f)

    js_data = load_data_js(data_js_file)

    # Construct new entry
    commit_info = bench_data.get("commit_info", {})

    entry = {
        "commit": {
            "time": commit_info.get("time"),
            "id": commit_info.get("id", "unknown"),
            "author": commit_info.get("author_name", "unknown"),
            "message": commit_info.get("message", "unknown"),
        },
        "date": int(time.time() * 1000),  # Current time in ms
        "tool": "casbin",
        "benchmarks": [],
    }

    for b in bench_data.get("benchmarks", []):
        name = b["name"]
        # Name is already normalized in format_benchmark_data.py

        # value is already in ns/op in format_benchmark_data.py
        value = b["stats"]["mean"]

        entry["benchmarks"].append({"name": name, "unit": "ns/op", "value": value})

    # Append to entries
    group_name = "Casbin"
    if group_name not in js_data["entries"]:
        js_data["entries"][group_name] = []

    js_data["entries"][group_name].append(entry)
    js_data["lastUpdate"] = int(time.time() * 1000)

    save_data_js(data_js_file, js_data)
    print(f"Updated {data_js_file} with {len(entry['benchmarks'])} benchmarks.")


if __name__ == "__main__":
    main()
