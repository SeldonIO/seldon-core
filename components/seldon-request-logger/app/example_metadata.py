metadata = {
    "request": {
        "type": "tabular",
        "schema": [{
            "name": "Age",
            "dtype": "float",
            "qdtype": "real"
        },
            {
            "name": "Workclass",
            "dtype": "int",
            "qdtype": "categorical",
            "n_categories": 9,
            "category_map": {
                "0": "?",
                "1": "Federal-gov",
                "2": "Local-gov",
                "3": "Never-worked",
                "4": "Private",
                "5": "Self-emp-inc",
                "6": "Self-emp-not-inc",
                "7": "State-gov",
                "8": "Without-pay"
            }
        },
            {
            "name": "Education",
            "dtype": "int",
            "qdtype": "categorical",
            "n_categories": 7,
            "category_map": {
                "0": "Associates",
                "1": "Bachelors",
                "2": "Doctorate",
                "3": "Dropout",
                "4": "High School grad",
                "5": "Masters",
                "6": "Prof-School"
            }
        },
            {
            "name": "Marital Status",
            "dtype": "int",
            "qdtype": "categorical",
            "n_categories": 4,
            "category_map": {
                "0": "Married",
                "1": "Never-Married",
                "2": "Separated",
                "3": "Widowed"
            }
        },
            {
            "name": "Occupation",
            "dtype": "int",
            "qdtype": "categorical",
            "n_categories": 9,
            "category_map": {
                "0": "?",
                "1": "Admin",
                "2": "Blue-Collar",
                "3": "Military",
                "4": "Other",
                "5": "Professional",
                "6": "Sales",
                "7": "Service",
                "8": "White-Collar"
            }
        },
            {
            "name": "Relationship",
            "dtype": "int",
            "qdtype": "categorical",
            "n_categories": 6,
            "category_map": {
                "0": "Husband",
                "1": "Not-in-family",
                "2": "Other-relative",
                "3": "Own-child",
                "4": "Unmarried",
                "5": "Wife"
            }
        },
            {
            "name": "Race",
            "dtype": "int",
            "qdtype": "categorical",
            "n_categories": 5,
            "category_map": {
                "0": "Amer-Indian-Eskimo",
                "1": "Asian-Pac-Islander",
                "2": "Black",
                "3": "Other",
                "4": "White"
            }
        },
            {
            "name": "Sex",
            "dtype": "int",
            "qdtype": "categorical",
            "n_categories": 2,
            "category_map": {
                "0": "Female",
                "1": "Male"
            }
        },
            {
            "name": "Capital Gain",
            "dtype": "float",
            "qdtype": "real"
        },
            {
            "name": "Capital Loss",
            "dtype": "float",
            "qdtype": "real"
        },
            {
            "name": "Hours per week",
            "dtype": "float",
            "qdtype": "real"
        },
            {
            "name": "Country",
            "dtype": "int",
            "qdtype": "categorical",
            "n_categories": 11,
            "category_map": {
                "0": "?",
                "1": "British-Commonwealth",
                "2": "China",
                "3": "Euro_1",
                "4": "Euro_2",
                "5": "Latin-America",
                "6": "Other",
                "7": "SE-Asia",
                "8": "South-America",
                "9": "United-States",
                "10": "Yugoslavia"
            }
        }
        ]
    },
    "response": {
        "type": "tabular",
        "schema": [{
            "name": "t:0",
            "dtype": "float",
            "qdtype": "real"
        }, {
            "name": "t:1",
            "dtype": "float",
            "qdtype": "real"
        }]
    }
}
