#define CATCH_CONFIG_MAIN
#include "catch_amalgamated.hpp"

#include <iostream>
#include <vector>
#include <memory>

#include "seldon/SeldonModel.hpp"

class TestModel : public seldon::SeldonModel {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        return data;
    }
};

TEST_CASE("TestSimpleMessageParsing", "Simple class parses data correctly") {

    std::cout << "Starting" << std::endl;

    TestModel tm = TestModel();
    std::cout << "Initialised" << std::endl;

    py::bytes input("{\"strData\":\"ndarray\"}");

    py::bytes result = tm.predictRaw(input);
    std::cout << "Ran predict raw" << std::endl;

    std::string resultString = result;
    std::cout << "result is " << resultString << std::endl;
}

