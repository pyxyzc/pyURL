package urltool

import "testing"

func TestGetBasePath(t *testing.T) {
	type args struct {
		targetUrl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "基本示例",
			args: args{
				targetUrl: "https://www.baidu.com/Go/",
			},
			want:    "Go",
			wantErr: false,
		},
		{
			name: "无效url示例",
			args: args{
				targetUrl: "xxx/Go/",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "空字符串",
			args: args{
				targetUrl: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "带query的url",
			args: args{
				targetUrl: "https://www.baidu.com/Go?name=1",
			},
			want:    "Go",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBasePath(tt.args.targetUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBasePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBasePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
