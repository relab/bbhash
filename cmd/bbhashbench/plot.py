import os
import re
import pandas as pd


def process_csv_file(plot_dir, filepath):
    df = pd.read_csv(filepath)
    grouped = df.groupby('Keys')

    # Compute average and std
    create_avg = grouped['CreateTime'].mean()
    create_std = grouped['CreateTime'].std()
    find_avg = grouped['FindTime'].mean()
    find_std = grouped['FindTime'].std()

    # Combine data into one DataFrame
    combined_create = pd.DataFrame({
        'Keys': create_avg.index,
        'CreateTime': create_avg.values,
        'CreateStd': create_std.values
    })

    combined_find = pd.DataFrame({
        'Keys': find_avg.index,
        'FindTime': find_avg.values,
        'FindStd': find_std.values
    })

    base_name = os.path.basename(filepath).replace('.csv', '')
    parent_dir = os.path.basename(os.path.dirname(filepath))

    os.makedirs(plot_dir, exist_ok=True)

    # Save to new CSV
    combined_create.to_csv(os.path.join(plot_dir, f"create_{base_name}_{parent_dir}.csv"), index=False)
    combined_find.to_csv(os.path.join(plot_dir, f"find_{base_name}_{parent_dir}.csv"), index=False)

    # Generate pgfplots code
    with open(os.path.join(plot_dir, f"{base_name}_{parent_dir}_pgfplots.tex"), 'w') as f:
        f.write("\\begin{figure}[ht!]\n\\centering\n")

        # Create plot
        f.write("\\begin{subfigure}[b]{0.45\\linewidth}\n")
        f.write(generate_pgfplots_code(f"create_{base_name}_{parent_dir}.csv", "CreateTime", "CreateStd"))
        f.write("\\end{subfigure}\n\\hfill\n")

        # Find plot
        f.write("\\begin{subfigure}[b]{0.45\\linewidth}\n")
        f.write(generate_pgfplots_code(f"find_{base_name}_{parent_dir}.csv", "FindTime", "FindStd"))
        f.write("\\end{subfigure}\n")

        f.write("\\end{figure}\n")


def pgfplots_normal(title, filename, avg_col, std_col):
    return f"""
\\begin{{tikzpicture}}
\\begin{{axis}}[
    title={{{title}}},
    xlabel=Keys,
    ylabel=Time (ms),
    xmode=log,
    ymode=log
]
\\addplot+[
    error bars/.cd,
    y dir=both,
    y explicit
] table [
    x=Keys,
    y={avg_col},
    y error={std_col},
    col sep=comma
] {{./{filename}}};
\\end{{axis}}
\\end{{tikzpicture}}
"""


def pgfplots_bar(title, filename, avg_col, std_col):
    return f"""
\\begin{{tikzpicture}}
\\begin{{axis}}[
    title={{{title}}},
    xlabel=Keys,
    ylabel=Time (ms),
    ymajorgrids=true,
    ybar,
    xmode=log,
    ymode=log,
    bar width=20pt
]
\\addplot[
    fill,
    error bars/.cd,
    y explicit
] table [
    x=Keys,
    y={avg_col},
    y error={std_col},
    col sep=comma
] {{{filename}}};
\\end{{axis}}
\\end{{tikzpicture}}
"""


def generate_pgfplots_code(filename, avg_col, std_col):
    # Extract necessary title components from the filename
    match = re.search(r'bbhash-(\w+)-gamma-(\d+\.\d+)-partitions-(\d+)', filename)
    if not match:
        raise ValueError(f"Unexpected filename format: {filename}")

    method, gamma, partitions = match.groups()

    method_name_mapping = {
        "seq": "Sequential",
        "par": "Parallel",
        "par2": "Parallel2"
    }
    action = "Create" if "Create" in avg_col else "Find"

    title_parts = [f"{action} {method_name_mapping.get(method, method)}", f"Gamma {gamma}"]
    if partitions != "1":
        title_parts.append(f"Partitions {partitions}")
    title = ", ".join(title_parts)
    return pgfplots_normal(title, filename, avg_col, std_col)


def generate_main_tex(plot_dir):
    tex_files = [f for f in os.listdir(plot_dir) if f.endswith('.tex')]

    # Sort the filenames
    def sort_key(filename):
        match = re.search(r'bbhash-(\w+)-gamma-(\d+\.\d+)-partitions-(\d+)', filename)
        if not match:
            return (1000, filename)  # Default high value for unknown formats
        method, _, partitions = match.groups()
        method_order = ["seq", "par", "par2"]
        return (int(partitions), method_order.index(method))

    tex_files.sort(key=sort_key)

    with open(os.path.join(plot_dir, 'main.tex'), 'w') as main_file:
        main_file.write('\\documentclass[11pt]{article}\n')
        main_file.write('\\usepackage{pgfplots}\n')
        main_file.write('\\pgfplotsset{compat=1.18}\n')
        main_file.write('\\usepackage{subcaption}\n')
        main_file.write('\\usepackage[margin=1in]{geometry}\n')  # Adjust page margins
        main_file.write('\\begin{document}\n')

        # Organize plots
        for idx, tex_file in enumerate(tex_files):
            main_file.write(f"\\input{{{tex_file}}}\n")
            main_file.write('\\vspace{1cm}\n')  # Add some vertical space between plots

        main_file.write('\\end{document}\n')


def main():
    start_dir = "bbhashbench_2023-09-15_21-49-14"
    plot_dir = "plot"
    print(f"Processing {start_dir}")

    for dirpath, _, filenames in os.walk(start_dir):
        for filename in filenames:
            if filename.endswith('.csv'):
                process_csv_file(plot_dir, os.path.join(dirpath, filename))

    generate_main_tex(plot_dir)


if __name__ == "__main__":
    main()
