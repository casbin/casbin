import pathlib, re, sys

try:
    p = pathlib.Path("comparison.md")
    if not p.exists():
        print("comparison.md not found, skipping post-processing.")
        sys.exit(0)

    lines = p.read_text(encoding="utf-8").splitlines()
    processed_lines = []
    in_code = False
    delta_col = None  # record "Diff" column start per table
    align_hint = None  # derived from benchstat header last pipe position

    ALIGN_COLUMN = 60  # fallback alignment when header not found

    def strip_worker_suffix(text: str) -> str:
        return re.sub(r'(\S+?)-\d+(\s|$)', r'\1\2', text)

    def get_icon(diff_val: float) -> str:
        if diff_val > 10:
            return "🐌"
        if diff_val < -10:
            return "🚀"
        return "➡️"

    def clean_superscripts(text: str) -> str:
        return re.sub(r'[¹²³⁴⁵⁶⁷⁸⁹⁰]', '', text)

    def parse_val(token: str):
        if '%' in token or '=' in token:
            return None
        token = clean_superscripts(token)
        token = token.split('±')[0].strip()
        token = token.split('(')[0].strip()
        if not token:
            return None

        m = re.match(r'^([-+]?\d*\.?\d+)([a-zA-Zµ]+)?$', token)
        if not m:
            return None
        try:
            val = float(m.group(1))
        except ValueError:
            return None
        suffix = (m.group(2) or "").replace("µ", "u")
        multipliers = {
            "n": 1e-9,
            "ns": 1e-9,
            "u": 1e-6,
            "us": 1e-6,
            "m": 1e-3,
            "ms": 1e-3,
            "s": 1.0,
            "k": 1e3,
            "K": 1e3,
            "M": 1e6,
            "G": 1e9,
            "Ki": 1024.0,
            "Mi": 1024.0**2,
            "Gi": 1024.0**3,
            "Ti": 1024.0**4,
            "B": 1.0,
            "B/op": 1.0,
            "C": 1.0,  # tolerate degree/unit markers that don't affect ratio
        }
        return val * multipliers.get(suffix, 1.0)

    def extract_two_numbers(tokens):
        found = []
        for t in tokens[1:]:  # skip name
            if t in {"±", "∞", "~", "│", "│"}:
                continue
            if '%' in t or '=' in t:
                continue
            val = parse_val(t)
            if val is not None:
                found.append(val)
                if len(found) == 2:
                    break
        return found

    # Pass 0: 
    # 1. find a header line with pipes to derive alignment hint
    # 2. calculate max content width to ensure right-most alignment
    max_content_width = 0
    
    for line in lines:
        if line.strip() == "```":
            in_code = not in_code
            continue
        if not in_code:
            continue
            
        # Skip footnotes/meta for width calculation
        if re.match(r'^\s*[¹²³⁴⁵⁶⁷⁸⁹⁰]', line) or re.search(r'need\s*>?=\s*\d+\s+samples', line):
            continue
        if not line.strip() or line.strip().startswith(('goos:', 'goarch:', 'pkg:', 'cpu:')):
            continue
        # Header lines are handled separately in Pass 1
        if '│' in line and ('vs base' in line or 'old' in line or 'new' in line):
            continue
            
        # It's likely a data line
        # Check if it has an existing percentage we might move/align
        curr_line = strip_worker_suffix(line).rstrip()
        pct_match = re.search(r'([+-]?\d+\.\d+)%', curr_line)
        if pct_match:
            # If we are going to realign this, we count width up to the percentage
            w = len(curr_line[:pct_match.start()].rstrip())
        else:
            w = len(curr_line)
        
        if w > max_content_width:
            max_content_width = w

    # Calculate global alignment target for Diff column
    # Ensure target column is beyond the longest line with some padding
    diff_col_start = max_content_width + 4
    
    # Calculate right boundary (pipe) position
    # Diff column width ~12 chars (e.g. "+100.00% 🚀")
    right_boundary = diff_col_start + 14

    for line in lines:

        if line.strip() == "```":
            in_code = not in_code
            processed_lines.append(line)
            continue

        if not in_code:
            processed_lines.append(line)
            continue

        # footnotes keep untouched
        if re.match(r'^\s*[¹²³⁴⁵⁶⁷⁸⁹⁰]', line) or re.search(r'need\s*>?=\s*\d+\s+samples', line):
            processed_lines.append(line)
            continue

        # header lines: ensure last column labeled Diff and force alignment
        if '│' in line and ('vs base' in line or 'old' in line or 'new' in line):
            # Strip trailing pipe and whitespace
            stripped_header = line.rstrip().rstrip('│').rstrip()
            
            # If "vs base" is present, ensure we don't duplicate "Diff" if it's already there
            # But we want to enforce OUR alignment, so we might strip existing Diff
            stripped_header = re.sub(r'\s+Diff\s*$', '', stripped_header, flags=re.IGNORECASE)
            stripped_header = re.sub(r'\s+Delta\b', '', stripped_header, flags=re.IGNORECASE)

            # Pad to diff_col_start
            padding = diff_col_start - len(stripped_header)
            if padding < 2: 
                padding = 2 # minimum spacing
                # If header is wider than data (unlikely but possible), adjust diff_col_start
                # But for now let's trust max_content_width or just append
            
            if len(stripped_header) < diff_col_start:
                new_header = stripped_header + " " * (diff_col_start - len(stripped_header))
            else:
                new_header = stripped_header + "  "

            # Add Diff column header if it's the second header row (vs base)
            if 'vs base' in line:
                new_header += "Diff"
            
            # Add closing pipe at the right boundary
            current_len = len(new_header)
            if current_len < right_boundary:
                new_header += " " * (right_boundary - current_len)
            
            new_header += "│"
            processed_lines.append(new_header)
            continue

        # non-data meta lines
        if not line.strip() or line.strip().startswith(('goos:', 'goarch:', 'pkg:')):
            processed_lines.append(line)
            continue

        original_line = line
        line = strip_worker_suffix(line)
        tokens = line.split()
        if not tokens:
            processed_lines.append(line)
            continue

        numbers = extract_two_numbers(tokens)
        pct_match = re.search(r'([+-]?\d+\.\d+)%', line)

        # Helper to align and append
        def append_aligned(left_part, content):
            if len(left_part) < diff_col_start:
                aligned = left_part + " " * (diff_col_start - len(left_part))
            else:
                aligned = left_part + "  "
            
            # Ensure content doesn't exceed right boundary (visual check only, we don't truncate)
            # But users asked not to exceed header pipe.
            # Header pipe is at right_boundary.
            # Content starts at diff_col_start.
            # So content length should be <= right_boundary - diff_col_start
            return f"{aligned}{content}"

        # Special handling for geomean when values missing or zero
        is_geomean = tokens[0] == "geomean"
        if is_geomean and (len(numbers) < 2 or any(v == 0 for v in numbers)) and not pct_match:
            leading = re.match(r'^\s*', line).group(0)
            left = f"{leading}geomean"
            processed_lines.append(append_aligned(left, "n/a (has zero)"))
            continue

        # when both values are zero, force diff = 0 and align
        if len(numbers) == 2 and numbers[0] == 0 and numbers[1] == 0:
            diff_val = 0.0
            icon = get_icon(diff_val)
            left = line.rstrip()
            processed_lines.append(append_aligned(left, f"{diff_val:+.2f}% {icon}"))
            continue

        # recompute diff when we have two numeric values
        if len(numbers) == 2 and numbers[0] != 0:
            diff_val = (numbers[1] - numbers[0]) / numbers[0] * 100
            icon = get_icon(diff_val)

            left = line
            if pct_match:
                left = line[:pct_match.start()].rstrip()
            else:
                left = line.rstrip()

            processed_lines.append(append_aligned(left, f"{diff_val:+.2f}% {icon}"))
            continue

        # fallback: align existing percentage to Diff column and (re)append icon
        if pct_match:
            try:
                pct_val = float(pct_match.group(1))
                icon = get_icon(pct_val)

                left = line[:pct_match.start()].rstrip()
                suffix = line[pct_match.end():]
                # Remove any existing icon after the percentage to avoid duplicates
                suffix = re.sub(r'\s*(🐌|🚀|➡️)', '', suffix)

                processed_lines.append(append_aligned(left, f"{pct_val:+.2f}% {icon}{suffix}"))
            except ValueError:
                processed_lines.append(line)
            continue

        # If we cannot parse numbers or percentages, keep the original (only worker suffix stripped)
        processed_lines.append(line)

    p.write_text("\n".join(processed_lines) + "\n", encoding="utf-8")

except Exception as e:
    print(f"Error post-processing comparison.md: {e}")
    sys.exit(1)
