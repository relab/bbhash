import os
import subprocess


def compile_go_program(go_file_path, output_path):
    env = os.environ.copy()
    env["GOOS"] = "linux"
    env["GOARCH"] = "amd64"

    cmd = ["go", "build", "-o", output_path, go_file_path]
    subprocess.check_call(cmd, env=env)


def distribute_binary(binary_path, user, machines):
    env = os.environ.copy()
    env["SHELL"] = "/bin/bash"
    for machine in machines:
        cmd = ["scp", binary_path, f"{user}@{machine}:~/"]
        subprocess.check_call(cmd)


def run_command_on_cluster(binary_name, user, machines):
    env = os.environ.copy()
    env["SHELL"] = "/bin/bash"
    for machine in machines:
        cmd = ["ssh", f"{user}@{machine}", f"./{binary_name}"]
        subprocess.check_call(cmd)


if __name__ == "__main__":
    go_file_path = "./main.go"
    binary_output_path = "./bbhashbench"
    user = os.getenv('USER')
    if not user:
        raise ValueError("USER environment variable is not set")

    machines = [f"bbchain{i}" for i in range(1, 31)]

    compile_go_program(go_file_path, binary_output_path)
    distribute_binary(binary_output_path, user, machines)
    run_command_on_cluster(binary_output_path.split('/')[-1], user, machines)
