#include "seldon/SeldonModel.hpp"

class ModelClass : public seldon::SeldonModelBase {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};

SELDON_DEFAULT_PYBIND_MODULE(ModelClass)
