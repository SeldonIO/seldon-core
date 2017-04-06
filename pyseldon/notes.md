How should the user go about deploying a model?

A model can be seen as a blackbox with several entry points. More specifically, it should be an object with some compulsory methods and some optional ones.

Compulsory methods:
	 - predict

Optional methods:
	 - predict_proba
	 - fit
	 - get_class_id_map
	 - score
	 - process_feedback

The entry point for predictions is predict_proba and default to predict if absent.

Use cases:

* Regression:
  * predict takes a list of lists (batch of points), returns a list of singleton containing the regression values
  * get_class_id_map returns a singleton that contains the name we want to give to the regression


* Classification (n classes):
  * predict_proba takes a list of lists (batch of poitns), returns a list of lists containing the class probailities for each point in the batch
  * get_class_id_map returns a list of class names

* Multi Regression:
  * predict returns list of lists
  * get_class_id_map returns list of names for each regression output

