# About

This Docker image provides a Python 3.8 environment with various data science and web development packages. It is designed to work with Dadosfera intelligence and transformation module, a platform for building, running, and deploying data science pipelines.

## Included packages

### Linux additions
- The chrome driver was added in the image as a dependency for webscraping tools like `selenium`

### Python packages
- awscli
- boto3
- snowflake-snowpark-python
- streamlit
- orchest
- dadosfera-library

### How to use this as a base image for an environment in our intelligence and processing module

1. Click on the `Environments` section
2. Use an existing environment or create a new environment
3. There are plenty of `Base Images`. Click on `+ Custom`
4. Type dadosfera/base-kernel-py:${VERSION}
5. Click on build to trigger the environment builds.

### How to add additional packages

In a new environment, new packages can be added using the following command:
```shell
mamba run -n py38 pip install $PACKAGE_NAME
```

### Setting up Selenium in the intelligence / transformation module


```python
from selenium import webdriver
from selenium.webdriver.chrome.options import Options

options = Options()
options.headless = True
options.add_argument('--no-sandbox')
options.add_argument('--disable-dev-shm-usage')
options.add_argument('--remote-debugging-port=9222')

# Provide the path to the Chrome WebDriver if it's not in the system's PATH
# driver = webdriver.Chrome(executable_path='/path/to/chromedriver', options=options)
driver = webdriver.Chrome(options=options)


# Navigate to a website
driver.get('https://www.example.com')

# Print the title of the website
print(driver.title)

# Close the browser
driver.quit()
```
