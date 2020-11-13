#define CATCH_CONFIG_MAIN
#include "catch_amalgamated.hpp"

#include <iostream>
#include <vector>

#include "seldon/SeldonModel.hpp"

class TestModel : public seldon::SeldonModel {

    seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) override {
        //data.has_data()
        google::protobuf::ListValue dataList = data.data().ndarray();
        std::vector<google::protobuf::Value> dataVec(dataList.values().cbegin(), dataList.values().cend());
        //std::cout << "DATA IS: " << dataVec[0].string_value() << "." << std::endl;
        std::cout << "DATA IS: " << data.strdata() << "." << std::endl;
        std::cout << "done fund." << std::endl;
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

