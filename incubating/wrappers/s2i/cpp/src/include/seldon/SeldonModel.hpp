#pragma once

#include <pybind11/pybind11.h>
#include <google/protobuf/util/json_util.h>

#include "prediction.pb.h"

namespace py = pybind11;

namespace seldon {

template <typename ProtoMessage = protos::SeldonMessage>
class SeldonModel
{
public:
    SeldonModel() { }

    virtual ~SeldonModel() { }

    virtual ProtoMessage predict(ProtoMessage &data) = 0;

    virtual py::bytes predictRaw(py::bytes &data) {

        py::buffer_info info(py::buffer(data).request());
        const char *charData = reinterpret_cast<const char *>(info.ptr);
        size_t charLength = static_cast<size_t>(info.size);

        std::string strData(charData, charLength);
        std::cout << "starting function with str: " << strData << std::endl;

        ProtoMessage input;
        google::protobuf::util::JsonStringToMessage(strData, &input);
        std::cout << "Converted to jsonmessage" << std::endl;

        ProtoMessage output = this->predict(input);

        std::cout << "Converting output" << std::endl;

        std::string outString;
        google::protobuf::util::MessageToJsonString(output, &outString);

        std::cout << "Returning output" << std::endl;
        return outString;
    }

};

using SeldonModelBase = SeldonModel<protos::SeldonMessage>;

#define SELDON_BIND_MODULE(PACKAGE, CLASS)  \
    PYBIND11_MODULE(PACKAGE, m)                       \
    {                                                 \
    py::class_<CLASS>(m, #CLASS)                      \
        .def(py::init())                              \
        .def("predict_raw", &CLASS::predictRaw); \
    }

#define SELDON_DEFAULT_BIND_MODULE()           \
    SELDON_BIND_MODULE(SeldonPackage, ModelClass)


}

