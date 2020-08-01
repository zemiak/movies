import { RestClientService } from "../../RestClientService.js";

export class MovieCsfdService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "MovieCsfdService";
    }

    getServicePath() {
        return "/movies/metadata/csfd/" + encodeURI(this.name);
    }

    setName(name) {
        this.name = name;
    }

    saveThumbnail(id, url, success, error) {
        fetch(this.getBaseUri() + "/movies/" + id + "/metadata/csfd",
             {method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify({url: url})})
            .then(response => success(response))
            .catch(err => error(err));
    }
}
