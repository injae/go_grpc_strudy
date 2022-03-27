window.onload = function() {
    //<editor-fold desc="Changeable Configuration Block">
    window.ui = SwaggerUIBundle({
        url: `http://0.0.0.0:8080/swagger.json`,
        dom_id: '#swagger-ui',
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

    //</editor-fold>
};
