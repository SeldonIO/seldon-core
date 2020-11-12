#pragma once

#include <pybind11/pybind11.h>

#include "seldon/proto/prediction.pb.h"

namespace py = pybind11;

namespace seldon {

template<typename ProtoClass = seldon::protos::SeldonMessage>
class SeldonModel 
{
public:
    SeldonModel();
    
    virtual ~SeldonModel();

    py::bytes& predictRaw(py::bytes &data);

    virtual ProtoClass& predict(ProtoClass &data) = 0;

private:

};

}

namespace seldon {

template<typename ProtoClass>
SeldonModel<ProtoClass>::SeldonModel() {

}

template<typename ProtoClass>
py::bytes& SeldonModel<ProtoClass>::predictRaw(py::bytes &data) {

}

}
