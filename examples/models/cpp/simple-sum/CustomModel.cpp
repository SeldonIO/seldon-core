#include <pybind11/pybind11.h>

#include "seldon/SeldonModel.hpp"

namespace py = pybind11;

class SeldonCustomModel : public seldon::SeldonModel {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};

PYBIND11_MODULE(seldon_custom_model, m) {
    py::class_<SeldonCustomModel>(m, "SeldonCustomModel")
        .def(py::init())
        .def("predict_raw", &SeldonCustomModel::predictRaw);
}

