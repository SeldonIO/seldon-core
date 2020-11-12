
#include "seldon/SeldonModel.hpp"

namespace seldon {

SeldonModel::SeldonModel() {  }

py::bytes& SeldonModel::predictRaw(py::bytes &data) {

    py::buffer_info info(py::buffer(data).request());
    const char *data = reinterpret_cast<const char *>(info.ptr);
    size_t length = static_cast<size_t>(info.size);

}

}
