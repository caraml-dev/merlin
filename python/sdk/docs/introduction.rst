.. _introduction:


***************
Introduction
***************

Merlin is a framework for serving machine learning model. The project was born of the belief that model deployment should be:

Easy and self-serve
  Human should not become the bottleneck for deploying model into production.

Scalable
  The model deployed should be able to handle Gojek scale and beyond.

Fast
  The framework should be able to let user iterate quickly.

Cost Efficient
  It should provide all benefit above in a cost efficient manner.

Merlin attempt to do so by:

**Abstracting Infrastructure** Merlin uses familiar concept such as Project, Model, and Version as its core component and abstract away complexity of deploying service from user.

**Auto Scaling** Merlin is built on top KNative and KFServing to provide a production ready serverless solution.








