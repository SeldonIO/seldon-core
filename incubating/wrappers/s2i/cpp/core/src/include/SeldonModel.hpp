#pragma once

#include <pybind11/pybind11.h>

#include "SeldonProto.hpp"

namespace py = pybind11;

namespace sc {

template<typename ProtoClass = seldon::protos::SeldonMessage>
class SeldonModel 
{
public:
    SeldonModel();

    py::bytes& predictRaw(py::bytes &data);

    virtual ProtoClass& predict(ProtoClass &data) = 0;

private:

};

}
