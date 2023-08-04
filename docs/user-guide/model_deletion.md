# Model Deletion

A Merlin model can be deleted only if it is not serving any endpoints and does not have any deployed model versions or, if the model is of the `pyfunc_v2` type, none of its model versions must not have any active prediction jobs. Deleting a model will result in the purging of all the model versions associated with it, as well as related entities such as endpoints or prediction jobs (applicable for models of the `pyfunc_v2` type) from the Merlin database. This action is **irreversible**.

A model with model versions that have any active prediction jobs or endpoints cannot be deleted.


## Model Deletion Via the SDK
To delete a Model, you can call the `delete_model()` function from the Merlin Python SDK.

```python
merlin.set_project("test-project")

merlin.set_model('test-model')

model = merlin.active_model()

model.delete_model()
```

## Model Deletion via the UI
To delete a model from the UI, you can access the delete button directly on the model list page. The dialog will provide information about any entities that are blocking the deletion process.

- If the model does not have any associated entities, a dialog like the one below will be displayed:
![Model Version Deletion Without Entity](../images/delete_model_no_entity.png)

- If the model has any associated active entities, a dialog like the one below will be displayed:
![Model Version Deletion Without Entity](../images/delete_model_active_entity.png)
