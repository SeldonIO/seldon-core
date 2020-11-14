#pragma once

#include <pybind11/pybind11.h>

#include "prediction.pb.h"

namespace py = pybind11;

namespace seldon {

class SeldonModel
{
public:
    SeldonModel();

    virtual ~SeldonModel();

    // TODO: Return without copy
    virtual protos::SeldonMessage predict(protos::SeldonMessage &data) = 0;

    // TODO: Return without copy
    py::bytes predictRaw(py::bytes &data);

};

}

