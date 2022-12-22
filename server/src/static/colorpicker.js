var defaultColors = [];
var colorList
var activeColor
var colorPicker


window.iro

function setColor(colorIndex) {
    // setActiveColor expects the color index!
    colorPicker.setActiveColor(colorIndex);
    }

function numColorSelect() {
    for (var i = 0; i < document.getElementById('numColorBox').value; i++) {
        defaultColors.push({ r: 255, g: 255, b: 255 });

        var original = document.getElementById("timeSelectDiv");
        var clone = original.cloneNode(true);
        clone.removeAttribute("id");
        clone.style.display = "inline";
        clone.className="timeSelectBox";
        document.getElementById("timeSelects").appendChild(clone); 
    }
    document.getElementById('numSelect').style.display = 'none';
    document.getElementById('wrap').style.display = 'inline';
    colorList = document.getElementById("colorList");
    activeColor = document.getElementById("activeColor");


    // Create a new color picker instance
    // https://iro.js.org/guide.html#getting-started
    colorPicker = new iro.ColorPicker(".colorPicker", {
        // color picker options
        // Option guide: https://iro.js.org/guide.html#color-picker-options
        width: 500,
        // Pure red, green and blue
        colors: defaultColors,
    
        handleRadius: 9,
        borderWidth: 1,
        borderColor: "#fff" });
    



    // https://iro.js.org/guide.html#color-picker-events
    colorPicker.on(["mount", "color:change"], function () {
    colorList.innerHTML = '';
    colorPicker.colors.forEach(color => {
        const index = color.index;
        const hexString = color.hexString;
        colorList.innerHTML += `
        <li onClick="setColor(${index})">
            <div class="swatch" style="background: ${hexString}"></div>
            <span>${index}: ${hexString}</span>
        </li>
        `;
    });
    });

    colorPicker.on(["mount", "color:setActive", "color:change"], function () {
    // colorPicker.color is always the active color
    const index = colorPicker.color.index;
    const hexString = colorPicker.color.hexString;
    activeColor.innerHTML = `
        <div class="swatch" style="background: ${hexString}"></div>
        <span>${index}: ${hexString}</span>
    `;
    });
}

function sendReq() {
    var timeSelects = document.getElementsByClassName("timeSelectBox");
    var times = [];
    for (var i = 0; i < timeSelects.length; i++) {
        times.push(timeSelects[i].children[0].value);
    }
    var colors = colorPicker.colors;
    var data = {Steps: []};
    for (var i = 0; i < colors.length; i++) {
        data.Steps.push({time: times[i]*500, color: colors[i]});
    }
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "update/hsvtime/"+document.getElementById("clientSelectBox").value, true);
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.send(JSON.stringify(data));
}
