
add_subdirectory(../../pybind11 ${CMAKE_CURRENT_BINARY_DIR}/pybind_build})

file(GLOB SELDON_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/*.cpp")

file(GLOB SELDON_PROTO_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/proto/*.cc")

add_library(
    seldon STATIC
    ${SELDON_SRC} ${SELDON_PROTO_SRC})

target_include_directories(
    seldon PUBLIC
    $<INSTALL_INTERFACE:include>
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
)

target_link_libraries(
    seldon PRIVATE 
    pybind11::module 
    pybind11::lto 
    pybind11::windows_extras)

pybind11_extension(seldon)
pybind11_strip(seldon)

set_target_properties(seldon PROPERTIES CXX_VISIBILITY_PRESET "hidden"
                                        CUDA_VISIBILITY_PRESET "hidden")

