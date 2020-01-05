import json
import sys

import yaml

TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link href="https://fonts.googleapis.com/css?family=Open+Sans:400,700|Source+Code+Pro:300,600|Titillium+Web:400,600,700" rel="stylesheet">
    <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.24.2/swagger-ui.css" >
    <style>
    html {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
    }
    *, *:before, *:after {
        box-sizing: inherit;
    }
    body {
      margin:0;
      background: #fafafa;
    }
    .markdown pre>code.language-json {
        font-family: consolas;
        font-style: italic;
    }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.20.0/swagger-ui-bundle.js"> </script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/3.20.0/swagger-ui-standalone-preset.js"> </script>
    <script>
    window.onload = function() {
        var spec = %s;
        window.ui = SwaggerUIBundle({
            spec: spec,
            dom_id: '#swagger-ui',
            validatorUrl: null,
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout"
        });
    }
    </script>
</body>
</html>
"""


def main():
    content = open(sys.argv[1], 'r', encoding='utf-8').read()
    spec = yaml.load(content, Loader=yaml.FullLoader)
    html = TEMPLATE % json.dumps(spec)
    with open(sys.argv[2], 'w') as f:
        f.write(html)


if __name__ == "__main__":
    main()

# python api_html.py main.go ./docs/api.yaml
