import { RestClientService } from "../RestClientService.js";

export class RootService extends RestClientService {
    constructor() {
        super();
    }

    getCustomEventName() {
        return "RootService";
    }

    getServicePath() {
        return "/ui/root";
    }
}
