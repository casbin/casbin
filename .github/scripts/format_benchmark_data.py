import json
import os
import sys
import datetime
import re


def normalize_name(name):
    # Remove "Benchmark" prefix
    name = re.sub(r"^Benchmark", "", name)
    # Remove -N suffix (GOMAXPROCS)
    name = re.sub(r"-\d+$", "", name)
    
    parts = re.split(r'[/_]', name)
    new_parts = []
    for p in parts:
        if p.lower() in ["rbac", "abac", "acl", "api", "rest"]:
            new_parts.append(p.upper())
        else:
            new_parts.append(p.capitalize())
    return "".join(new_parts)


def main():
    if len(sys.argv) < 3:
        print("Usage: python format_benchmark_data.py input.txt output.json")
        sys.exit(1)

    input_path = sys.argv[1]
    output_path = sys.argv[2]

    # Get commit info from environment variables
    # These should be set in the GitHub Action
    commit_info = {
        "author": {
            "email": os.environ.get("COMMIT_AUTHOR_EMAIL", ""),
            "name": os.environ.get("COMMIT_AUTHOR_NAME", ""),
            "username": os.environ.get("COMMIT_AUTHOR_USERNAME", ""),
        },
        "committer": {
            "email": os.environ.get("COMMIT_COMMITTER_EMAIL", ""),
            "name": os.environ.get("COMMIT_COMMITTER_NAME", ""),
            "username": os.environ.get("COMMIT_COMMITTER_USERNAME", ""),
        },
        "distinct": True,  # Assuming true for push to master
        "id": os.environ.get("COMMIT_ID", ""),
        "message": os.environ.get("COMMIT_MESSAGE", ""),
        "timestamp": os.environ.get("COMMIT_TIMESTAMP", ""),
        "tree_id": os.environ.get("COMMIT_TREE_ID", ""),
        "url": os.environ.get("COMMIT_URL", ""),
    }

    # Get CPU count
    cpu_count = os.cpu_count() or 1

    benches = []
    
    try:
        with open(input_path, "r", encoding="utf-8") as f:
            lines = f.readlines()
            
        for line in lines:
            # Parse Go benchmark output: BenchmarkName-8 10000 123 ns/op
            # Also handle lines with 4 columns if any, but standard is: Name Iterations Value Unit
            match = re.search(r'^(Benchmark\S+)\s+(\d+)\s+([\d\.]+)\s+ns/op', line)
            if match:
                name = match.group(1)
                iterations = int(match.group(2))
                val_ns = float(match.group(3))
                
                # Format extra info
                extra = f"{iterations} times"
                
                # Create entry
                benches.append({
                    "name": normalize_name(name),
                    "value": round(val_ns, 2),
                    "unit": "ns/op",
                    "extra": extra
                })
                
    except Exception as e:
        print(f"Error processing {input_path}: {e}")
        sys.exit(1)

    output_data = {
        "commit": commit_info,
        "date": int(datetime.datetime.now().timestamp() * 1000),  # Current timestamp in ms
        "tool": "go",
        "procs": cpu_count,
        "benches": benches,
    }

    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(output_data, f, indent=2)

    print(f"Successfully formatted benchmark data to {output_path}")


if __name__ == "__main__":
    main()
