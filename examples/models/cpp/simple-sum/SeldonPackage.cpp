#include <pybind11/pybind11.h>

#include "seldon/SeldonModel.hpp"

namespace py = pybind11;

class ModelClass : public seldon::SeldonModel {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};

PYBIND11_MODULE(SeldonPackage, m) {
    py::class_<ModelClass>(m, "ModelClass")
        .def(py::init())
        .def("predict_raw", &ModelClass::predictRaw);
}

