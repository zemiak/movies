import { RestClientService } from "../RestClientService.js";

export class MovieDetailService extends RestClientService {
    constructor() {
        super();
        this.id = null;
    }

    setId(id) {
        this.id = id;
    }

    getCustomEventName() {
        return "MovieDetailService";
    }

    getServicePath() {
        return "/movies/detail/" + this.id;
    }

    saveOrUpdate(item, success, error) {
        var url = "/movies"
        var methodName = null === item.id ? "POST" : "PUT";

        fetch(this.getBaseUri() + url, {method: methodName, headers: {"Content-Type": "application/json"}, body: JSON.stringify(item)})
            .then(response => success(response))
            .catch(err => error(err));
    }
}
