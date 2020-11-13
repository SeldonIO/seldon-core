#define CATCH_CONFIG_MAIN
#include "catch_amalgamated.hpp"

#include <iostream>

#include "seldon/SeldonModel.hpp"

class TestModel : public seldon::SeldonModel {

    seldon::protos::SeldonMessage& predict(seldon::protos::SeldonMessage &data) override {
        std::cout << "DATA IS: " << data.ByteSize() << "." << std::endl;
    }
};

TEST_CASE("TestSimpleMessageParsing", "Simple class parses data correctly") {
    TestModel tm = TestModel();
    py::bytes input("{}");
    py::bytes result = tm.predictRaw(input);

    std::string resultString = result;
    std::cout << "result is " << resultString << std::endl;
}

