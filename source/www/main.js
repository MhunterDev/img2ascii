document.addEventListener("DOMContentLoaded", function () {
    var form = document.getElementById("uploadForm");
    var submit = document.getElementById("submit");
    var asciiOutput = document.getElementById("asciiOutput");
    var aspectMode = document.getElementById("aspectMode");
    var sizeOptions = document.getElementById("sizeOptions");

    // Handle aspect mode changes
    if (aspectMode && sizeOptions) {
        aspectMode.addEventListener("change", function() {
            if (aspectMode.value === "fixed") {
                sizeOptions.style.display = "block";
            } else {
                sizeOptions.style.display = "none";
            }
        });
    }

    if (form && submit && asciiOutput) {
        form.addEventListener("submit", function (event) {
            event.preventDefault();
            submit.disabled = true;
            asciiOutput.textContent = "Uploading...";
            var formData = new FormData(form);
            fetch("/upload", {
                method: "POST",
                body: formData
            })
            .then(response => {
                if (!response.ok) throw new Error(response.statusText);
                return response.text();
            })
            .then(text => {
                asciiOutput.textContent = text;
            })
            .catch(error => {
                asciiOutput.textContent = "Error: " + error.message;
            })
            .finally(() => {
                submit.disabled = false;
            });
        });
    }

    var bannerForm = document.getElementById("bannerGen");
    var bannerSubmit = document.getElementById("bannerSubmit");
    if (bannerForm && bannerSubmit && asciiOutput) {
        bannerForm.addEventListener("submit", function (event) {
            event.preventDefault();
            bannerSubmit.disabled = true;
            asciiOutput.textContent = "Generating banner...";
            var formData = new FormData(bannerForm);
            fetch("/banner", {
                method: "POST",
                body: formData
            })
            .then(response => {
                if (!response.ok) throw new Error(response.statusText);
                return response.text();
            })
            .then(text => {
                asciiOutput.textContent = text;
            })
            .catch(error => {
                asciiOutput.textContent = "Error: " + error.message;
            })
            .finally(() => {
                bannerSubmit.disabled = false;
            });
        });
    }
});