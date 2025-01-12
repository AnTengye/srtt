package google

import "testing"

func Test_translateText(t *testing.T) {
	type args struct {
		targetLanguage string
		text           string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ja-zh",
			args: args{
				targetLanguage: "zh",
				text:           "ちょっと、父さんに代わろう。",
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := translateText(tt.args.targetLanguage, tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("translateTextBasic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("translateTextBasic() got = %v, want %v", got, tt.want)
			}
		})
	}
}
