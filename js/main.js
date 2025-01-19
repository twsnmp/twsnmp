
//===============================================================
// debounce関数
//===============================================================
function debounce(func, wait) {
    var timeout;
    return function() {
        var context = this, args = arguments;
        var later = function() {
            timeout = null;
            func.apply(context, args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}


//===============================================================
// メニュー関連
//===============================================================

// 変数でセレクタを管理
var $menubar = $('#menubar');
var $menubarHdr = $('#menubar_hdr');

// menu
$(window).on("load resize", debounce(function() {
    if(window.innerWidth <= 900) {	// ここがブレイクポイント指定箇所です
        // 小さな端末用の処理
        $('body').addClass('small-screen').removeClass('large-screen');
        $menubar.addClass('display-none').removeClass('display-block');
        $menubarHdr.removeClass('display-none ham').addClass('display-block');
    } else {
        // 大きな端末用の処理
        $('body').addClass('large-screen').removeClass('small-screen');
        $menubar.addClass('display-block').removeClass('display-none');
        $menubarHdr.removeClass('display-block').addClass('display-none');

        // ドロップダウンメニューが開いていれば、それを閉じる
        $('.ddmenu_parent > ul').hide();
    }
}, 10));

$(function() {

    // ハンバーガーメニューをクリックした際の処理
    $menubarHdr.click(function() {
        $(this).toggleClass('ham');
        if ($(this).hasClass('ham')) {
            $menubar.addClass('display-block');
        } else {
            $menubar.removeClass('display-block');
        }
    });

    // アンカーリンクの場合にメニューを閉じる処理
    $menubar.find('a[href*="#"]').click(function() {
        $menubar.removeClass('display-block');
        $menubarHdr.removeClass('ham');
    });

    // ドロップダウンの親liタグ（空のリンクを持つaタグのデフォルト動作を防止）
	$menubar.find('a[href=""]').click(function() {
		return false;
	});

	// ドロップダウンメニューの処理
    $menubar.find('li:has(ul)').addClass('ddmenu_parent');
    $('.ddmenu_parent > a').addClass('ddmenu');

// タッチ開始位置を格納する変数
var touchStartY = 0;

// タッチデバイス用
$('.ddmenu').on('touchstart', function(e) {
    // タッチ開始位置を記録
    touchStartY = e.originalEvent.touches[0].clientY;
}).on('touchend', function(e) {
    // タッチ終了時の位置を取得
    var touchEndY = e.originalEvent.changedTouches[0].clientY;
    
    // タッチ開始位置とタッチ終了位置の差分を計算
    var touchDifference = touchStartY - touchEndY;
    
    // スクロール動作でない（差分が小さい）場合にのみドロップダウンを制御
    if (Math.abs(touchDifference) < 10) { // 10px以下の移動ならタップとみなす
        var $nextUl = $(this).next('ul');
        if ($nextUl.is(':visible')) {
            $nextUl.stop().hide();
        } else {
            $nextUl.stop().show();
        }
        $('.ddmenu').not(this).next('ul').hide();
        return false; // ドロップダウンのリンクがフォローされるのを防ぐ
    }
});

    //PC用
    $('.ddmenu_parent').hover(function() {
        $(this).children('ul').stop().show();
    }, function() {
        $(this).children('ul').stop().hide();
    });

    // ドロップダウンをページ内リンクで使った場合に、ドロップダウンを閉じる
    $('.ddmenu_parent ul a').click(function() {
        $('.ddmenu_parent > ul').hide();
    });

});


//===============================================================
// 小さなメニューが開いている際のみ、body要素のスクロールを禁止。
//===============================================================
$(function() {
  function toggleBodyScroll() {
    // 条件をチェック
    if ($('#menubar_hdr').hasClass('ham') && !$('#menubar_hdr').hasClass('display-none')) {
      // #menubar_hdr が 'ham' クラスを持ち、かつ 'display-none' クラスを持たない場合、スクロールを禁止
      $('body').css({
        overflow: 'hidden',
        height: '100%'
      });
    } else {
      // その他の場合、スクロールを再び可能に
      $('body').css({
        overflow: '',
        height: ''
      });
    }
  }

  // 初期ロード時にチェックを実行
  toggleBodyScroll();

  // クラスが動的に変更されることを想定して、MutationObserverを使用
  const observer = new MutationObserver(toggleBodyScroll);
  observer.observe(document.getElementById('menubar_hdr'), { attributes: true, attributeFilter: ['class'] });
});


//===============================================================
// スムーススクロール（※バージョン2024-7）※通常タイプ
//===============================================================
$(function() {
    // ページトップボタンの表示・非表示
    var scroll = $('.pagetop');
    var scrollShow = $('.pagetop-show');
    $(scroll).hide();
    $(window).scroll(function() {
        if ($(this).scrollTop() >= 300) {
            $(scroll).fadeIn().addClass(scrollShow);
        } else {
            $(scroll).fadeOut().removeClass(scrollShow);
        }
    });

    // スムーススクロール
    function smoothScroll(target) {
        var scroll = target.offset().top ; // スクロール位置を調整
        $('body,html').animate({ scrollTop: scroll }, 500);
    }

    // ページトップボタンのクリックイベント
    $('.pagetop').click(function(e) {
        e.preventDefault();
        $('body,html').animate({ scrollTop: 0 }, 500);
    });

    // ページ読み込み時のハッシュ処理
    $(window).on('load', function() {
        var hash = location.hash;
        if (hash) {
            $('body,html').scrollTop(0);
            setTimeout(function() {
                var target = $(hash);
                if (target.length) {
                    smoothScroll(target);
                }
            }, 100);
        }
    });

    // リンククリック時のスムーススクロール
    $(window).on('load', function() {
        $('a[href*="#"]').click(function(e) {
            var href = $(this).attr('href');
            var targetId = href.split('#')[1]; // ハッシュ部分だけを取り出す
            var target = $('#' + targetId);
            if (target.length) {
                e.preventDefault();
                smoothScroll(target);
                history.pushState(null, null, '#' + targetId); // ハッシュをURLに追加
            }
        });
    });
});


//===============================================================
// 汎用開閉処理
//===============================================================
$(function() {
	$('.openclose').next().hide();
	$('.openclose').click(function() {
		$(this).next().slideToggle();
		$('.openclose').not(this).next().slideUp();
	});
});
