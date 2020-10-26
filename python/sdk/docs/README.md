# Documentation

## Building locally

Install Python requirements:
```bash
pip install -r requirements_docs.txt
```

Install `pandoc`
```bash
conda install -c conda-forge pandoc=1.19.2
```

Generate doc
```bash
make html
```

The output is located at `_build` directory

Re-generating doc (delete all the files in the `_build` directory before generating the documentation again)
```bash
make clean
```

## Publishing

The documentation for the latest **commit** is located [here](https://merlin-sdk.readthedocs.io/en/latest/).
