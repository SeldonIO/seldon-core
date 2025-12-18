# How to run the Jupyter notebook

## Create a virtual environment and use it in Jupyter Lab

1. Create a virtual environment
```bash
python -m venv xgboost-env
```

2. Activate it
```bash
source xgboost-env/bin/activate
```

3. Install the Jupyter kernel using the Python interpreter in this environment
```bash
python -m ipykernel install --user \
  --name xgboost-env \
  --display-name "Python (xgboost-env)"
```

4. Install development dependencies
```bash
pip install -r requirements-dev.txt
```

5. Additionally, install the `seldon_core` package
  * You can either install it with pip from pypi with `pip install seldon-core`
  * Or you can install it from source by:
    *  Navigating to the <root>/python folder and `make install`

6. Install the package dependencies
```bash
pip install -r ./xgboostserver/requirements.txt
```

## Open Jupyter Lab and run the cells

1. Open Jupyter Lab
```bash
jupyter lab
```

2. Navigate to the browser, then to the `test` folder and select the notebook

3. From the top right, choose the Kernel called `Python (xgboost-env)`

4. Run the cells one after the other