import json
from typing import List, Dict, Union
import numpy as np


class NumpyEncoder(json.JSONEncoder):
    def default(self, obj):  # pylint: disable=arguments-differ,method-hidden
        """
        Encode Numpy Arrays as JSON
        Parameters
        ----------
        obj
             JSON Encoder

        """
        if isinstance(
            obj,
            (
                np.int_,
                np.intc,
                np.intp,
                np.int8,
                np.int16,
                np.int32,
                np.int64,
                np.uint8,
                np.uint16,
                np.uint32,
                np.uint64,
            ),
        ):
            return int(obj)
        elif isinstance(obj, (np.float_, np.float16, np.float32, np.float64)):
            return float(obj)
        elif isinstance(obj, (np.ndarray,)):
            return obj.tolist()
        return json.JSONEncoder.default(self, obj)


def read_inputs_as_numpy(inputs: Union[List, Dict]) -> np.ndarray:
    """
    Read payload inputs as np.ndarray enforcing float32/int32 dtypes.
    See: https://github.com/SeldonIO/seldon-core/issues/3940 for details.
    """
    x = np.array(inputs)
    if x.dtype == np.float64:
        return x.astype(np.float32)
    elif x.dtype == np.int64:
        return x.astype(np.int32)
    else:
        return x
