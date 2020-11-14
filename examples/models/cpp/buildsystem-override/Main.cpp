#include "seldon/SeldonModel.hpp"

class MyModelClass : public seldon::SeldonModelBase {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};

SELDON_BIND_MODULE(CustomSeldonPackage, MyModelClass)
