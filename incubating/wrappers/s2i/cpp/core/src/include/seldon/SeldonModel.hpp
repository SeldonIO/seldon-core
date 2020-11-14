#pragma once

#include <pybind11/pybind11.h>
#include <google/protobuf/util/json_util.h>

#include <iostream>
#include <memory>

#include "prediction.pb.h"

namespace py = pybind11;

namespace seldon {

class SeldonModel 
{
public:
    SeldonModel() {}
    
    virtual ~SeldonModel() {}

    // TODO: Return without copy
    py::bytes predictRaw(py::bytes &data) {

        py::buffer_info info(py::buffer(data).request());
        const char *charData = reinterpret_cast<const char *>(info.ptr);
        size_t charLength = static_cast<size_t>(info.size);

        std::string strData(charData, charLength);
        std::cout << "starting function with str: " << strData << std::endl;

        seldon::protos::SeldonMessage input;
        google::protobuf::util::JsonStringToMessage(strData, &input);
        std::cout << "Converted to jsonmessage" << std::endl;

        seldon::protos::SeldonMessage output = this->predict(input);

        std::cout << "Converting output" << std::endl;

        std::string outString;
        google::protobuf::util::MessageToJsonString(output, &outString);

        std::cout << "Returning output" << std::endl;
        return outString;
    }


    virtual seldon::protos::SeldonMessage predict(seldon::protos::SeldonMessage &data) = 0;

};

}

