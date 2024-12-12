# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import re
import sys
import math
import matplotlib.pyplot as plt


def parse_benchmark_output(filename):
    """
    Parses the Go benchmark output and extracts the relevant data.
    Returns a dictionary with the structure:
    {
        operation: {
            distribution: {
                'sizes': [size1, size2, ...],
                'times': [time1, time2, ...],
            }
        }
    }
    """
    benchmark_data = {}
    with open(filename, "r") as f:
        for line in f:
            # Adjust the regex to match your benchmark output format
            match = re.match(
                r"^BenchmarkTrieOperations/(\w+)_size_(\d+)/(\w+)-\d+\s+"  # Benchmark name
                r"\s*(\d+)\s+"  # Iterations
                r"\s*([\d\,\.]+)\s+ns/op",  # Time per op
                line,
            )
            if match:
                distribution, size, operation, iterations, time_per_op = match.groups()
                size = int(size)
                time_per_op = float(
                    time_per_op.replace(",", "")
                )  # Remove comma if present

                # Initialize nested dictionaries if necessary
                if operation not in benchmark_data:
                    benchmark_data[operation] = {}
                if distribution not in benchmark_data[operation]:
                    benchmark_data[operation][distribution] = {"sizes": [], "times": []}

                benchmark_data[operation][distribution]["sizes"].append(size)
                benchmark_data[operation][distribution]["times"].append(time_per_op)
            else:
                # Optionally, print or log the lines that did not match to debug
                # print(f"Line did not match: {line.strip()}")
                pass  # In production, you might want to handle this differently
    return benchmark_data


def compute_expected_times(benchmark_data):
    """
    Computes expected times for each operation based on O(log n),
    scaling from the average actual time at n = 1000.
    Returns a dictionary with the expected sizes and times per operation.
    """
    expected_data = {}
    sizes = sorted(
        {
            size
            for op_data in benchmark_data.values()
            for dist_data in op_data.values()
            for size in dist_data["sizes"]
        }
    )
    base_log = math.log2(1000)  # Changed baseline to n=1000
    for operation in benchmark_data.keys():
        # Collect actual times at n = 1000 across distributions
        actual_times_at_base = []
        for distribution in benchmark_data[operation].keys():
            dist_data = benchmark_data[operation][distribution]
            for size, time in zip(dist_data["sizes"], dist_data["times"]):
                if size == 1000:
                    actual_times_at_base.append(time)
        if not actual_times_at_base:
            continue  # Skip if no data at baseline size

        # Compute average actual time at baseline size
        base_time = sum(actual_times_at_base) / len(actual_times_at_base)

        # Compute expected times for all sizes
        expected_sizes = sizes
        expected_times = [base_time * (math.log2(n) / base_log) for n in expected_sizes]

        expected_data[operation] = {
            "sizes": expected_sizes,
            "times": expected_times,
        }
    return expected_data


def plot_combined_benchmarks(benchmark_data, expected_data=None):
    """
    Plots combined benchmarks for each operation, with lines for each distribution.
    """
    for operation, distributions in benchmark_data.items():
        plt.figure(figsize=(10, 6))
        # For each key distribution
        sizes = []
        for distribution, data in distributions.items():
            sizes.extend(data["sizes"])
        sizes = sorted(set(sizes))
        for distribution, data in distributions.items():
            dist_sizes = data["sizes"]
            times = data["times"]
            # Sort sizes and times together based on sizes
            dist_sizes, times = zip(*sorted(zip(dist_sizes, times)))
            plt.plot(dist_sizes, times, label=distribution, marker="o")

        # Plot expected data if provided
        if expected_data and operation in expected_data:
            expected_sizes = expected_data[operation]["sizes"]
            expected_times = expected_data[operation]["times"]
            plt.plot(
                expected_sizes,
                expected_times,
                label="Expected (O(log n))",
                linestyle="--",
                marker="x",
                color="black",
            )

        plt.title(f"HAMT {operation.capitalize()} Benchmark")
        plt.xlabel("Size (n)")
        plt.ylabel("Time per operation (ns)")
        plt.xscale("log")
        plt.xticks(sizes, sizes, rotation=45)
        plt.legend()
        plt.grid(True, which="both", ls="--", alpha=0.5)
        plt.tight_layout()
        plt.savefig(f"hamt_{operation}_benchmark.png")
        plt.close()


def main():
    if len(sys.argv) < 2:
        print(
            "Usage: python extract_and_graph_hamt_benchmarks.py <benchmark_output_file>"
        )
        sys.exit(1)

    benchmark_output_file = sys.argv[1]
    benchmark_data = parse_benchmark_output(benchmark_output_file)

    # Compute expected times based on average actual times at baseline size (e.g., n=1000)
    expected_data = compute_expected_times(benchmark_data)

    if not expected_data:
        print(
            "Warning: Expected data could not be computed because baseline size data is missing."
        )
        print("Please ensure that the benchmark includes data for the baseline size.")
    else:
        plot_combined_benchmarks(benchmark_data, expected_data)


if __name__ == "__main__":
    main()
