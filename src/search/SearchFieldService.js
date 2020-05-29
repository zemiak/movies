import { RestClientService } from "../RestClientService.js";

class SearchFieldService {
    constructor() {
        this.searchField = document.getElementById("searchField");
        this.searchField.onkeypress = ev => this.keyPress(ev);

        this.searchButton = document.getElementById("searchButton");
        this.searchButton.onclick = ev => this.search(ev);
    }

    keyPress(e) {
        if (13 === e.keyCode) {
            this.search();
        }
    }

    search() {
        var text = this.searchField.value;
        if (text.length > 2) {
            document.location = "/search/" + encodeURIComponent(text);
        } else {
            console.log("SearchFieldService.search(): text too short: " + text);
        }
    }
}

export const searchFieldService = new SearchFieldService();
