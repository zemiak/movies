import { RestClientService } from "../RestClientService.js";

export class SerieDetailService extends RestClientService {
    constructor() {
        super();
        this.id = null;
    }

    setId(id) {
        this.id = id;
    }

    getCustomEventName() {
        return "SerieDetailService";
    }

    getServicePath() {
        return "/series/detail/" + this.id;
    }

    saveOrUpdate(item, success, error) {
        var url = "/series"
        var methodName = null === item.id ? "POST" : "PUT";

        fetch(this.getBaseUri() + url, {method: methodName, headers: {"Content-Type": "application/json"}, body: JSON.stringify(item)})
            .then(response => success(response))
            .catch(err => error(err));
    }
}
