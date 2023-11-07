ERR_DRY_RUN = """
Some module might be missing from your pyfunc model.
If the module is part of your code, 
you can specify it as 'code' argument when calling log_pyfunc, e.g.:

merlin.log_pyfunc_model(model_instance=MyModel(), 
        conda_env="env.yaml",
        code=["path/to/your/package"])
        
If the module is a python package, then you should add it in the
conda env.yaml that you pass to log_pyfunc method, e.g.:

dependencies:
- python=3.8
- pip:
  - my-missing-package==1.0
  
OR if the package is not available in PyPI distribution

dependencies:
- python=3.8
- my-missing-package=1.0

See: https://docs.conda.io/projects/conda/en/latest/user-guide/tasks/manage-environments.html#creating-an-environment-from-an-environment-yml-file
"""
