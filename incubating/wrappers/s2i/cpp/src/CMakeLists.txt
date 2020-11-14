
find_package(Protobuf REQUIRED)
find_package(PythonLibs REQUIRED)

add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../pybind11 ${CMAKE_CURRENT_BINARY_DIR}/pybind_build})

if(SELDON_OPT_BUILD_PROTO)
    execute_process(COMMAND make -C ${CMAKE_CURRENT_SOURCE_DIR}/../
 proto-build)
endif()
if(SELDON_OPT_CLONE_PYBIND11)
    execute_process(COMMAND make -C ${CMAKE_CURRENT_SOURCE_DIR}/../
 clone-pybind)
endif()

file(GLOB SELDON_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/*.cpp")

file(GLOB SELDON_PROTO_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/proto/*.cc")

message(STATUS "printing")
message(STATUS ${SELDON_SRC})
message(STATUS ${SELDON_PROTO_SRC})
message(STATUS "printed")

if(SELDON_OPT_BUILD_STATIC)
    add_library(
        seldon STATIC
        ${SELDON_PROTO_SRC}
        ${SELDON_SRC})
else()
    add_library(
        seldon SHARED
        ${SELDON_PROTO_SRC}
        ${SELDON_SRC})
endif()

target_include_directories(
    seldon PUBLIC
    $<INSTALL_INTERFACE:include>
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
)

target_include_directories(
    seldon PUBLIC 
    ${PYTHON_INCLUDE_DIRS})

target_link_libraries(
    seldon PRIVATE
    pybind11::module
    pybind11::lto
    ${PROTOBUF_LIBRARIES})

if(MSVC)
    target_link_libraries(
        seldon PRIVATE 
        pybind11::windows_extras)
endif()

add_library(seldon::seldon ALIAS seldon)

if(SELDON_OPT_INSTALL)
    install(TARGETS seldon EXPORT SeldonTargets
        LIBRARY DESTINATION lib
        ARCHIVE DESTINATION lib
        INCLUDES DESTINATION include)

    install(DIRECTORY include/ DESTINATION include)

    install(DIRECTORY ${PROJECT_SOURCE_DIR}/src/include
        DESTINATION include)

    install(EXPORT SeldonTargets
        FILE seldon-config.cmake
        NAMESPACE seldon::
        DESTINATION lib/cmake/seldon)
endif()


