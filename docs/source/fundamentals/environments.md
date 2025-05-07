(environments)=

# Environments

```{eval-rst}
.. meta::
   :description: This page contains information about how to create and manage environments in Dadosfera AI.
```

Environments are the foundation of your {term}`pipelines <(data science) pipeline>` in Dadosfera AI. They define the Python packages and system dependencies that are available to your pipeline steps.

```{figure} ../img/environments-list.png
:align: center
:width: 768
:alt: List of environments for a given Dadosfera AI project

The list of environments for a given Dadosfera AI project.
```

(environment-creation)=

## Creating an environment

To create a new environment, follow these instructions:

1. Click on _Environments_ in the navigation bar.
2. Click the _+ new environment_ button to create a new environment.
3. Configure the environment.
4. Press _create environment_.

```{figure} ../img/environment-creation.png
:align: center
:width: 768
:alt: Creating a new environment in Dadosfera AI

Creating a new environment in Dadosfera AI.
```

(environment-configuration)=

## Environment configuration

When creating an environment, you can configure the following:

- **Name**: A unique name for your environment.
- **Base Image**: The base Docker image to use for your environment. You can choose from a list of pre-built images or specify a custom image.
- **Python Version**: The Python version to use in your environment.
- **System Dependencies**: System packages to install in your environment.
- **Python Packages**: Python packages to install in your environment.

```{figure} ../img/environment-configuration.png
:align: center
:width: 768
:alt: Configuring a new environment in Dadosfera AI

Configuring a new environment in Dadosfera AI.
```

(environment-management)=

## Managing environments

You can manage your environments in the following ways:

- **Edit**: Click on the environment you want to edit and make your changes.
- **Delete**: Click on the environment you want to delete and press the _delete_ button.
- **Duplicate**: Click on the environment you want to duplicate and press the _duplicate_ button.

```{figure} ../img/environment-management.png
:align: center
:width: 768
:alt: Managing environments in Dadosfera AI

Managing environments in Dadosfera AI.
```

(environment-usage)=

## Using environments in pipelines

To use an environment in a pipeline, follow these instructions:

1. Open the pipeline you want to edit.
2. Click on the step you want to configure.
3. In the _Environment_ dropdown, select the environment you want to use.
4. Press _Save_.

```{figure} ../img/environment-usage.png
:align: center
:width: 768
:alt: Using environments in pipelines in Dadosfera AI

Using environments in pipelines in Dadosfera AI.
```

(environment-best-practices)=

## Best practices

Here are some best practices for working with environments:

- **Use specific versions**: Always specify the exact version of Python packages you want to use. This ensures reproducibility.
- **Keep environments small**: Only install the packages you need. This makes your environments faster to build and run.
- **Use base images**: Use pre-built base images when possible. This saves time and ensures compatibility.
- **Test environments**: Test your environments before using them in production pipelines.

```{figure} ../img/environment-best-practices.png
:align: center
:width: 768
:alt: Best practices for working with environments in Dadosfera AI

Best practices for working with environments in Dadosfera AI.
```
