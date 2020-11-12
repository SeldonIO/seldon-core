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

    py::bytes& predictRaw(py::bytes &data);

    virtual seldon::protos::SeldonMessage& predict(seldon::protos::SeldonMessage &data) = 0;

private:

};

}

