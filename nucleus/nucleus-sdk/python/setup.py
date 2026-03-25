from setuptools import setup

setup(
    name="nucleus-sdk",
    version="1.0.0",
    description="Python SDK for the NUCLEUS AI-native shell runtime",
    py_modules=["nucleus"],
    python_requires=">=3.8",
    extras_require={
        "streaming": ["websocket-client>=1.0"],
    },
)
