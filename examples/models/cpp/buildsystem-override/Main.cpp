#include <pybind11/pybind11.h>

#include "seldon/SeldonModel.hpp"

namespace py = pybind11;

class MyModelClass : public seldon::SeldonModelBase {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};

PYBIND11_MODULE(CustomSeldonPackage, m) {
    py::class_<MyModelClass>(m, "MyModelClass")
        .def(py::init())
        .def("predict_raw", &MyModelClass::predictRaw);
}

