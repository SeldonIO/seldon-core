from flask import Flask, request

app = Flask(__name__)

@app.route("/", methods=['GET','POST'])
def index():
    try:
        content = request.get_json(force=True)
        print(str(content))
        return str(content)
    except:
        return 'Error processing input'


if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)