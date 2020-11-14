include(FindProtobuf)
find_package(Protobuf REQUIRED)

file(GLOB SELDON_SRC
    "${PROJECT_SOURCE_DIR}/src/*.cpp")

file(GLOB SELDON_PROTO_SRC
    "${PROJECT_SOURCE_DIR}/src/proto/*.cc")

add_executable(seldon-test
    Main.cpp catch_amalgamated.cpp
    ${SELDON_SRC} ${SELDON_PROTO_SRC}
    )

message(STATUS ${PROTOBUF_LIBRARY})

#pybind11_add_module(
#    seldon-test
#    Main.cpp catch_amalgamated.cpp)

#set_property(TARGET seldon-test PROPERTY ENABLE_EXPORTS 1)

target_include_directories(
    seldon-test PUBLIC
    $<INSTALL_INTERFACE:include>
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>)

target_include_directories(
    seldon-test PUBLIC
    ${PROJECT_SOURCE_DIR}/src/include)

message(STATUS ${PROJECT_SOURCE_DIR}/src/include)

include_directories(${PYTHON_INCLUDE_DIRS})

target_link_libraries(
    seldon-test PRIVATE
    pybind11::module
    pybind11::lto
    ${PROTOBUF_LIBRARIES}
    ${PYTHON_LIBRARIES})

    #pybind11_extension(seldon-test)
    #pybind11_strip(seldon-test)
    #
    #set_target_properties(
    #    seldon-test PROPERTIES 
    #        CXX_VISIBILITY_PRESET "hidden"
    #        CUDA_VISIBILITY_PRESET "hidden")



