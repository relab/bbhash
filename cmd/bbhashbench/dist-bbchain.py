import os
import subprocess
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor


def compile_go_program(go_file_path, output_path):
    env = os.environ.copy()
    env["GOOS"] = "linux"
    env["GOARCH"] = "amd64"

    cmd = ["go", "build", "-o", output_path, go_file_path]
    subprocess.check_call(cmd, env=env)


def dist_to_machine(binary_path, user, machine):
    cmd = ["scp", binary_path, f"{user}@{machine}:~/"]
    subprocess.check_call(cmd)


def distribute_binary(binary_path, user, machines):
    with ThreadPoolExecutor() as executor:
        futures = [executor.submit(dist_to_machine, binary_path, user, machine) for machine in machines]
        for future in futures:
            future.result()  # waits for thread to complete and raises exceptions if any occurred


def fetch_csv_files(user, machines):
    # Create a directory named by the current date and time
    date_str = datetime.now().strftime('%Y-%m-%d_%H-%M-%S')
    base_directory = f"./bbhashbench_{date_str}"
    os.makedirs(base_directory, exist_ok=True)

    for machine in machines:
        # Create a subdirectory for each machine
        machine_directory = os.path.join(base_directory, machine)
        os.makedirs(machine_directory, exist_ok=True)

        remote_path = f"{user}@{machine}:~/*.csv"
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
    go_file_path = "./main.go"
    binary_output_path = "./bbhashbench"
    binary = binary_output_path.split('/')[-1]
    user = os.getenv('USER')
    if not user:
        raise ValueError("USER environment variable is not set")

    machines = [f"bbchain{i}" for i in range(2, 30)]
    compile_go_program(go_file_path, binary_output_path)
    distribute_binary(binary_output_path, user, machines)

    common = ["-gamma", "2.0", "-count", "20"]
    configurations = []
    configurations.append((machines[0], common + ["-name", "seq"]))
    configurations.append((machines[1], common + ["-name", "seq"]))
    configurations.append((machines[2], common + ["-name", "par"]))
    configurations.append((machines[3], common + ["-name", "par"]))
    machine_index = 4
    partitions = [2, 4, 6, 8, 10, 12, 16, 24, 28, 32, 64, 128]
    for partition in partitions:
        args = common + ["-name", "par2", "-partitions", str(partition)]
        configurations.append((machines[machine_index], args))
        machine_index += 1
        configurations.append((machines[machine_index], args))
        machine_index += 1

    parallel_run(binary, user, configurations)

    fetch_csv_files(user, machines)
