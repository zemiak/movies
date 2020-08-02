import { RestClientService } from "../../RestClientService.js";

export class MovieCsfdService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "MovieCsfdService";
    }

    getServicePath() {
        return "/metadata/csfd/" + encodeURI(this.name);
    }

    setName(name) {
        this.name = name;
    }

    saveThumbnail(id, data, success, error) {
        fetch(this.getBaseUri() + "/metadata/csfd/" + id,
             {method: "PUT", headers: {"Content-Type": "application/json"}, body: JSON.stringify(data)})
            .then(response => success(response))
            .catch(err => error(err));
    }
}
