import json
import sys
import glob


def merge_jsons(output_file, input_files):
    merged_data = None
    all_benchmarks = []

    for file_path in input_files:
        with open(file_path, "r", encoding="utf-8") as f:
            data = json.load(f)
            if merged_data is None:
                merged_data = data
            all_benchmarks.extend(data.get("benchmarks", []))

    if merged_data:
        merged_data["benchmarks"] = all_benchmarks
        with open(output_file, "w", encoding="utf-8") as f:
            json.dump(merged_data, f, indent=4)


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python merge_benchmarks.py output.json input1.json input2.json ...")
        sys.exit(1)

    output_file = sys.argv[1]
    input_files = sys.argv[2:]

    # Expand globs if shell didn't
    expanded_inputs = []
    for p in input_files:
        expanded_inputs.extend(glob.glob(p))

    merge_jsons(output_file, expanded_inputs)
