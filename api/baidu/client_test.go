package baidu

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func TestClient_Translate(t *testing.T) {
	client := resty.New().SetBaseURL(defaultUrl)
	logger := zap.NewNop().Sugar()
	type fields struct {
		apiKey  string
		secret  string
		httpCli *resty.Client
		logger  *zap.SugaredLogger
	}
	type args struct {
		text       string
		sourceLang string
		targetLang string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				apiKey:  "",
				secret:  "",
				httpCli: client,
				logger:  logger,
			},
			args: args{
				text:       "何恥ずかしがってんだ、白って言え。",
				sourceLang: "ja",
				targetLang: "zh",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				apiKey:  tt.fields.apiKey,
				secret:  tt.fields.secret,
				httpCli: tt.fields.httpCli,
				logger:  tt.fields.logger,
			}
			got, err := c.Translate(tt.args.text, tt.args.sourceLang, tt.args.targetLang)
			if (err != nil) != tt.wantErr {
				t.Errorf("Translate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == "" {
				t.Errorf("Translate() got = %v, want not %v", got, tt.want)
			}
		})
	}
}
