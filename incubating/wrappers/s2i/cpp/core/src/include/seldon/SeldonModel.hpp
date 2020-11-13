#pragma once

#include <pybind11/pybind11.h>
#include <google/protobuf/util/json_util.h>

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

        seldon::protos::SeldonMessage input;
        google::protobuf::util::JsonStringToMessage(strData, &input);

        seldon::protos::SeldonMessage output = this->predict(output);

        std::string outString;
        google::protobuf::util::MessageToJsonString(output, &outString);

        return outString;
    }


    virtual seldon::protos::SeldonMessage& predict(seldon::protos::SeldonMessage &data) = 0;

};

}

