# Dadosfera Base Kernel Python Agent

This Dockerfile builds a custom Docker image for the Dadosfera platform, providing a Python-based environment tailored for data science and processing pipelines. It includes dependencies for Jupyter, AWS integration, and Dadosfera-specific tools.


## Features

- **Base Image**: Built on top of `jupyter/base-notebook:2022-03-09`.
- **Enterprise Gateway Support**: Includes kernel files and dependencies for Jupyter Enterprise Gateway.
- **Python Environment**: Pre-installed Python libraries for data science, AWS integration, and Dadosfera-specific tools.
- **Goose CLI**: Debugging and configuration using the Goose CLI.
- **AWS Integration**: Supports AWS credentials and services like CodeArtifact.
- **Custom Configuration**: Environment variables and configurations for Jupyter and Orchest.

## How to use it:

Just add `dadosfera/base-kernel-py-agent:<version>`, with `<version>` being e.g. `1.0.0`, as a custom image inside a Project's Environment, and build it:

![image](https://github.com/user-attachments/assets/97ed5bea-69ed-4585-9f46-f00cb724ff8c)


## Included Tools and Libraries

### System Dependencies
- `openssh-server`, `git`, `default-libmysqlclient-dev`, `libkrb5-dev`
- `bzip2`, `libxcb1`, `libdbus-1-3`, `curl`, `ca-certificates`

### Python Libraries
- `boto3`, `awscli`, `chardet`, `requests`, `numpy<2`, `pandas`
- `snowflake-snowpark-python`, `fastparquet`, `streamlit`
- Dadosfera-specific libraries: `dadosfera==1.8.0b6`, `dadosfera_logs==1.0.3`

### Additional Tools
- Goose CLI for debugging and configuration.
- Jupyter Enterprise Gateway kernel files.

## Environment Variables

- `PATH`: Includes `/root/.local/bin` for Goose CLI.
- `JUPYTER_PATH`: Ensures the correct Jupyter executable is used.
- `HOME`: Set to `/home/$NB_USER`.
- `BASH_ENV`: Points to `.orchestrc` for shell initialization.
- `CONDA_ENV`: Specifies the default Conda environment (`base`).
- `PLOTLY_RENDERER`: Default renderer for Plotly in JupyterLab (`iframe`).
- `KERNEL_LANGUAGE`: Specifies the kernel language (`python`).
- `ORCHEST_VERSION`: Dynamically set during the build process.

## Build Instructions

To build the Docker image, run the following command:

```bash
docker build -t dadosfera-base-kernel-py-agent .
```

## Usage

This image is designed to be used as a base image for Dadosfera's intelligence and processing modules. It supports running Jupyter kernels and integrates seamlessly with AWS services.

## Customization

- Modify the `requirements-user.txt` and `requirements.txt` files to include additional Python dependencies.
- Update the `bootscript.sh` file to customize the container's startup behavior.

## Notes

- Ensure that AWS credentials are provided as secrets during the build process.
- The image is optimized to avoid cache busting by copying application files late in the build process.
