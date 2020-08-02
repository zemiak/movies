import { RestClientService } from "../../RestClientService.js";

export class MovieItunesService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "MovieItunesService";
    }

    getServicePath() {
        return "/metadata/itunes/" + encodeURI(this.name);
    }

    setName(name) {
        this.name = name;
    }

    saveThumbnail(id, url, success, error) {
        fetch(this.getBaseUri() + "/metadata/itunes/" + id,
             {method: "PUT", headers: {"Content-Type": "application/json"}, body: JSON.stringify({url: url})})
            .then(response => success(response))
            .catch(err => error(err));
    }
}
