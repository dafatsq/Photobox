export namespace admin {
	
	export class Frame {
	    id: string;
	    label: string;
	    filePath: string;
	    template: string;
	
	    static createFrom(source: any = {}) {
	        return new Frame(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.filePath = source["filePath"];
	        this.template = source["template"];
	    }
	}

}

