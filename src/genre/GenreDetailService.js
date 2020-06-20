import { RestClientService } from "../RestClientService.js";

export class GenreDetailService extends RestClientService {
    constructor() {
        super();
        this.id = null;
    }

    setId(id) {
        this.id = id;
    }

    getCustomEventName() {
        return "GenreDetailService";
    }

    getServicePath() {
        return "/genres/" + this.id;
    }

    saveOrUpdate(item, success, error) {
        var url = "/genres"
        var methodName = null === item.id ? "POST" : "PUT";

        fetch(this.getBaseUri() + url, {method: methodName, headers: {"Content-Type": "application/json"}, body: JSON.stringify(item)})
            .then(response => success(response))
            .catch(err => error(err));
    }
}
