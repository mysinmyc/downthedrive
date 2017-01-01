package httpd

import (
	"html/template"
	"net/http"

	"fmt"

	"net/url"

	"github.com/mysinmyc/downthedrive"
	"github.com/mysinmyc/gocommons/diagnostic"
)

const (
	TEMPLATE_BROWSEPAGE = `
<html>
<head>
<link rel="stylesheet" type="text/css" href="/style.css">
<link rel="stylesheet" href="http://fontawesome.io/assets/font-awesome/css/font-awesome.css">

</head>
<body>


{{ if .IsFolder }}

	<div class="topBar"><i class="fa fa-folder-o fa-2" aria-hidden="true"></i>&nbsp{{.Name}}</div>

	<div class="content">
		<table>


			{{if .ParentReference}}
				<tr><td><a href="/browse?id={{.ParentReference.Id}}" class="fileElement">..</td></tr>
			{{end}}

			{{range .Children}}
				<tr>
					<td>
						<a href="/browse?id={{.Id}}" class="fileElement">
						
						{{ if .Folder }}
							<i class="fa fa-folder-o fa-1" aria-hidden="true"></i>&nbsp
						{{end}}
						{{.Name}}</a>
					</td>
					<td>{{.LastModifiedDateTime}}</td>
				</tr>
			{{end}}
		</table>
	</div>
{{else}}
	<table style="width:100%;height:100%">
		
		<tr><td  class="topBar">{{.Name}} &nbsp; <a href="{{if .File}}/download?id={{.Id}}&contentType={{.File.MimeType}}&fileName={{.Name}}{{else}}/download?id={{.Id}}&fileName={{.Name}}{{end}}">
		</a>
		</tr></td>
	</table>
{{end}}

</body>
</html>
	`
)

func RegisterBrowseDriveHandler(pDownTheDrive *downthedrive.DownTheDrive) error {

	vTemplate, vTemplateError := template.New("browse_html").Parse(TEMPLATE_BROWSEPAGE)

	if vTemplateError != nil {
		return vTemplateError
	}
	http.Handle("/browse", &BrowseDriveHandler{ItemHandlerBase: ItemHandlerBase{downTheDrive: pDownTheDrive}, template: vTemplate})

	return nil

}

type BrowseDriveHandler struct {
	ItemHandlerBase
	template *template.Template
}

func (vSelf *BrowseDriveHandler) ServeHTTP(pResponse http.ResponseWriter, pRequest *http.Request) {

	vItem, vError := vSelf.getRequestedItem(pRequest)

	if vError != nil {
		diagnostic.LogWarning("BrowseDriveHandler.ServeHTTP", "an error occurred", diagnostic.NewError("failed to  failed to process request ", vError))
		http.Error(pResponse, fmt.Sprintf("%v", vError), 500)
	}

	if vItem == nil {
		diagnostic.LogWarning("BrowseDriveHandler.ServeHTTP", "item not found %s", nil, vItem)
		http.NotFound(pResponse, pRequest)
		return
	}

	if vItem.IsFile() {
		http.Redirect(pResponse, pRequest, fmt.Sprintf("/download?id=%s&fileName=%s", vItem.Id, url.QueryEscape(vItem.Name)), 302)
		return
	}

	//vItemForTemplate := &ItemForTemplate{vItem}
	vError = vSelf.template.Execute(pResponse, vItem)

	if vError != nil {
		diagnostic.LogWarning("BrowseDriveHandler.ServeHTTP", "an error occurred", diagnostic.NewError("failed to  failed to process request ", vError))
		http.Error(pResponse, fmt.Sprintf("%v", vError), 500)
	}
}
