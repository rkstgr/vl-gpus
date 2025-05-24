build on mac
```sh
go build -o gpu-metrics-collector main.go
```

build for linux
```sh
GOOS=linux GOARCH=amd64 go build -o gpu-metrics-collector main.go
```

# Local testing
Make the script executable
```
chmod +x fake-nvidia-smi.sh
```

Test it works
```sh
./fake-nvidia-smi.sh --query-gpu=index,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw --format=csv,noheader,nounits
```

Move it so its available in PATH
```sh
cp fake-nvidia-smi.sh ~/.local/bin/nvidia-smi
chmod +x ~/.local/bin/nvidia-smi
```

Verify it's found
```sh
which nvidia-smi
nvidia-smi --version  # Should show our fake output
```
