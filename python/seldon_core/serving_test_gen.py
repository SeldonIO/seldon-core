"""Contains methods to generate a JSON file for Seldon API integration testing."""

import os
from typing import List, Optional, Union

import numpy as np
import pandas as pd

RANGE_INTEGER_MIN = 0
RANGE_INTEGER_MAX = 1
RANGE_FLOAT_MIN = 0.0
RANGE_FLOAT_MAX = 1.0


def _column_range(col: pd.Series) -> Optional[List]:
    """
    Calculate minimum and maximum of a column and outputs a list.

    Parameters
    ----------
    col
        Column to inspect.

    Returns
    -------
        Min and max of the column range as a list.
    """
    if col.dtype == np.float:
        if pd.isnull(min(col)):  # This also means that maximum is null
            return [RANGE_FLOAT_MIN, RANGE_FLOAT_MAX]
        else:
            return [min(col), max(col)]
    elif col.dtype == np.integer:
        if pd.isnull(min(col)):  # This also means that maximum is null
            return [RANGE_INTEGER_MIN, RANGE_INTEGER_MAX]
        else:
            return [min(col), max(col)]
    else:
        return np.NaN


def _column_values(column: pd.Series) -> Union[List, float]:
    """
    Create a list of unique values for categorical variables.

    Parameters
    ----------
    column
        Column to inspect.

    Returns
    -------
        List of unique values for categorical variables
    """
    if column.dtype != np.number:
        return column.unique().tolist()
    else:
        return np.NaN


def create_seldon_api_testing_file(
    data: pd.DataFrame, target: str, output_path: str
) -> bool:
    """
    Create a JSON file for Seldon API testing.

    Parameters
    ----------
    data
        Pandas DataFrame used as a recipe for the json file.
    target
        Name of the target column.
    output_path
        Path of output file.

    Returns
    -------
        True if file correctly generated.
    """

    # create a Data frame in the form of JSON object
    df_for_json = pd.DataFrame(data=data.columns.values, columns=["name"])
    df_for_json["dtype"] = np.where(
        data.dtypes == np.float,
        "FLOAT",
        np.where(data.dtypes == np.int, "INTEGER", np.NaN),
    )
    df_for_json["ftype"] = np.where(
        data.dtypes == np.number, "continuous", "categorical"
    )
    ranges = [_column_range(data[column_name]) for column_name in data.columns.values]
    values = [_column_values(data[column_name]) for column_name in data.columns.values]
    df_for_json["range"] = ranges
    df_for_json["values"] = values
    # Split the target
    df_for_json_target = df_for_json[df_for_json.name == target]
    df_for_json_features = df_for_json[df_for_json.name != target]

    # Convert data frames to JSON with a trick that removes records with NaNs
    json_features_df = df_for_json_features.T.apply(
        lambda row: row[~row.isnull()].to_json()
    )
    json_features = f'[{",".join(json_features_df)}]'
    json_target_df = df_for_json_target.T.apply(
        lambda row: row[~row.isnull()].to_json()
    )
    json_target = f'[{",".join(json_target_df)}]'
    json_combined = f'{{"features": {json_features}, "targets": {json_target}}}'

    with open(output_path, "w+") as output_file:
        output_file.write(str(json_combined))
    return os.path.exists(output_path)
