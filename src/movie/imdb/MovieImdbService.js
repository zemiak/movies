import { RestClientService } from "../../RestClientService.js";

export class MovieImdbService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "MovieImdbService";
    }

    getServicePath() {
        return "/metadata/imdb/" + encodeURI(this.name);
    }

    setName(name) {
        this.name = name;
    }

    saveThumbnail(id, data, success, error) {
        fetch(this.getBaseUri() + "/metadata/imdb/" + id,
             {method: "PUT", headers: {"Content-Type": "application/json"}, body: JSON.stringify(data)})
            .then(response => success(response))
            .catch(err => error(err));
    }
}
