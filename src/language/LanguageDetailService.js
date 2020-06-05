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
}
