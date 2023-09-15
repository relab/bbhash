import os
import subprocess


def compile_go_program(go_file_path, output_path):
    env = os.environ.copy()
    env["GOOS"] = "linux"
    env["GOARCH"] = "amd64"

    cmd = ["go", "build", "-o", output_path, go_file_path]
    subprocess.check_call(cmd, env=env)


def distribute_binary(binary_path, user, machines):
    for machine in machines:
        cmd = ["scp", binary_path, f"{user}@{machine}:~/"]
        subprocess.check_call(cmd)


def run(binary_name, user, machines, args=[]):
    for machine in machines:
        remote_command = f"./{binary_name} {' '.join(args)}"
        cmd = ["ssh", f"{user}@{machine}", remote_command]
        subprocess.check_call(cmd)


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
    run(binary, user, machines[0:2], common + ["-name", "seq"])
    run(binary, user, machines[2:4], common + ["-name", "par"])

    partitions = [2, 4, 6, 8, 10, 12, 16, 24, 28, 32, 64, 128]
    machine_index = 4  # Starting index for the machines list
    for partition in partitions:
        args = common + ["-name", "par2", "-partitions", str(partition)]
        run(binary, user, machines[machine_index:machine_index+2], args)
        machine_index += 2

# 2, 4, 6, 8, 10, 12, 16, 24, 28, 32, 48, 64, 128}
