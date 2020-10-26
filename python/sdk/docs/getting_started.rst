.. _getting_started:


***************
Getting started
***************

.. _installing-sdk:

Installation
============


Install `merlin-sdk` from PyPI by running this command:
::
    pip install merlin-sdk

Authenticate to `gcloud`:
::
    gcloud auth application-default login

Now you are ready to deploy your machine learning model. See sample notebook to get you started.

.. _concept:


Concept
========

Project
    Project represent a namespace for a collection of model.
    In subsequent Merlin release, Project would be the main building block for access control.

Model
    Model represent machine learning model.
    A model can have a type which determine how the model can be deployed.
    Merlin supports both standard model (XGBoost, SKLearn, Tensorflow, and PyTorch) and user-defined model (PyFunc model).
    Conceptually, model in Merlin is similar to a class in programming language.
    To instantiate a model you’ll have to create a model version.

Model Version
    Model version represents a snapshot of particular model iteration. A model version might contain artifacts which is deployable to Merlin.
    Each of a model version will have a version endpoint. Merlin supports up to 3 model version to be deployed at the same time.

Version Endpoint
    Version endpoint is URL associated with a model version deployment.
    Version endpoint URL has following template
    ::

        http://<model_name>-<version>.<project_name>.<merlin_base_url>

    For example a model named mymodel within project named myproject will
    have a version endpoint for version 1 which look as follow:

    ::

        http://mymodel-1.myproject.models.id.merlin.dev

    Version endpoint has several state:

    pending
        Is the initial state of a version endpoint.
    ready
        Once deployed, a version endpoint is in ready state and is accessible.
    serving
        A version endpoint is in serving state if Model Endpoint has traffic rule which uses the particular Version Endpoint.
        A model version could not be undeployed if its version endpoint is in serving state.
    terminated
        Once undeployed a version endpoint is in terminated state.
    failed
        If error occurred during deployment.

Model Endpoint
    Model Endpoint is a stable URL associated with a model.
    Model endpoint URL has following template
    ::

        http://<model_name>.<project_name>.<merlin_base_url>

    For example a model named mymodel within project named myproject will
    have model endpoint which look as follow:

    ::

        http://mymodel.myproject.models.id.merlin.dev


    Model endpoint can have a traffic rule which determine which model version will receive traffic when request is received.
