//#include <google/protobuf/util/json_util.h>
//#include <string>
//
//#include "seldon/SeldonModel.hpp"
//
//namespace seldon {
//
//SeldonModel::~SeldonModel() {  }
//
//SeldonModel::SeldonModel() {  }
//
//py::bytes SeldonModel::predictRaw(py::bytes &data) {
//
//    py::buffer_info info(py::buffer(data).request());
//    const char *charData = reinterpret_cast<const char *>(info.ptr);
//    size_t charLength = static_cast<size_t>(info.size);
//
//    std::string strData(charData, charLength);
//
//    seldon::protos::SeldonMessage input;
//    google::protobuf::util::JsonStringToMessage(strData, &input);
//
//    seldon::protos::SeldonMessage output = this->predict(output);
//
//    std::string outString;
//    google::protobuf::util::MessageToJsonString(output, &outString);
//
//    return outString;
//}
//
//}
