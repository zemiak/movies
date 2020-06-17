import { RestClientService } from "../RestClientService.js";

export class LanguageDetailService extends RestClientService {
    constructor() {
        super();
        this.id = null;
    }

    setId(id) {
        this.id = id;
    }

    getCustomEventName() {
        return "LanguageDetailService";
    }

    getServicePath() {
        return "/languages/" + this.id;
    }

    saveOrUpdate(item, success, error) {
        var url = "/languages/10"
        var methodName = "" === item.id ? "POST" : "PUT";

        fetch(this.getBaseUri() + url, {method: methodName, headers: {"Content-Type": "application/json"}, body: JSON.stringify(item)})
            .then(response => success(response))
            .catch(err => error(err));
    }
}
