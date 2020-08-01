import { RestClientService } from "../../RestClientService.js";

export class MovieImdbService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "MovieImdbService";
    }

    getServicePath() {
        return "/movies/metadata/imdb/" + encodeURI(this.name);
    }

    setName(name) {
        this.name = name;
    }

    saveThumbnail(id, url, success, error) {
        fetch(this.getBaseUri() + "/movies/" + id + "/metadata/imdb",
             {method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify({url: url})})
            .then(response => success(response))
            .catch(err => error(err));
    }
}
