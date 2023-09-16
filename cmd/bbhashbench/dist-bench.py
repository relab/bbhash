import os
import subprocess
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor


def compile_go_test():
    cmd = ["go", "test", "-c", "./"]
    subprocess.check_call(cmd)


def dist_to_machine(binary_path, user, machine):
    cmd = ["ssh", f"{user}@{machine}", "pkill", "-f", binary_path]
    # Don't check_call here because it will fail if the process is not running
    subprocess.call(cmd)
    cmd = ["scp", binary_path, f"{user}@{machine}:~/"]
    subprocess.check_call(cmd)


def distribute_binary(binary_path, user, machines):
    with ThreadPoolExecutor() as executor:
        futures = [executor.submit(dist_to_machine, binary_path, user, machine) for machine in machines]
        for future in futures:
            future.result()  # waits for thread to complete and raises exceptions if any occurred


def fetch_bench_files(user, machines):
    # Create a directory named by the current date and time
    date_str = datetime.now().strftime('%Y-%m-%d_%H-%M-%S')
    base_directory = f"./bbhash.test_{date_str}"
    os.makedirs(base_directory, exist_ok=True)

    for machine in machines:
        # Create a subdirectory for each machine
        machine_directory = os.path.join(base_directory, machine)
        os.makedirs(machine_directory, exist_ok=True)

        remote_path = f"{user}@{machine}:~/*.txt"
        cmd = ["scp", remote_path, machine_directory]
        subprocess.check_call(cmd)


def run(binary_name, user, machine, args=[]):
    remote_command = f"./{binary_name} {' '.join(args)}"
    cmd = ["ssh", f"{user}@{machine}", remote_command]
    subprocess.check_call(cmd)


def parallel_run(binary_name, user, configurations):
    with ThreadPoolExecutor() as executor:
        futures = [executor.submit(run, binary_name, user, machine, args) for machine, args in configurations]
        for future in futures:
            future.result()  # waits for thread to complete and raises exceptions if any occurred


if __name__ == "__main__":
    binary = "bbhash.test"
    user = os.getenv('USER')
    if not user:
        raise ValueError("USER environment variable is not set")

    machines = [f"bbchain{i}" for i in range(2, 30)]
    compile_go_test()
    distribute_binary(binary, user, machines)

    common = ["-test.run=none",  "-test.count=1", "-test.timeout=0"]
    configurations = []
    for machine in machines[:15]:
        configurations.append((machine, common + ["-test.bench=BenchmarkNewBBHash"]))
    for machine in machines[15:]:
        configurations.append((machine, common + ["-test.bench=BenchmarkFind"]))

    parallel_run(binary, user, configurations)

    fetch_bench_files(user, machines)
